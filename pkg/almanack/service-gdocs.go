//go:build goexperiment.rangefunc

package almanack

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/iterx"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (svc Services) ConfigureGoogleCert(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	if svc.Gsvc.HasCert() {
		return nil
	}

	opt, err := svc.Queries.GetOption(ctx, "google-json")
	switch {
	case db.IsNotFound(err):
		l := almlog.FromContext(ctx)
		l.Warn("ConfigureGoogleCert: no certificate in database")
		return nil
	case err != nil:
		return err
	default:
		return svc.Gsvc.ConfigureCert(opt)
	}
}

func (svc Services) CreateGDocsDoc(ctx context.Context, externalID string) (dbDoc *db.GDocsDoc, err error) {
	defer errorx.Trace(&err)

	if err := svc.ConfigureGoogleCert(ctx); err != nil {
		return nil, err
	}
	cl, err := svc.Gsvc.GDocsClient(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := gdocs.Request(ctx, cl, externalID)
	if err != nil {
		// TODO: figure out common errors, like no-permissions
		return nil, err
	}

	newDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
		ExternalID: externalID,
		Document:   *doc,
	})
	if err != nil {
		return nil, err
	}
	return &newDoc, nil
}

func (svc Services) ProcessGDocs(ctx context.Context) error {
	docs, err := svc.Queries.ListGDocsWhereUnprocessed(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, doc := range docs {
		docErr := svc.ProcessGDocsDoc(ctx, doc)
		errs = append(errs, docErr)
	}
	return errors.Join(errs...)
}

var ascii = bytemap.Range(0, 127)

func (svc Services) ProcessGDocsDoc(ctx context.Context, dbDoc db.GDocsDoc) (err error) {
	defer errorx.Trace(&err)

	// Get existing image uploads
	rows, err := svc.Queries.ListGDocsImagesByExternalID(ctx, dbDoc.ExternalID)
	if err != nil {
		return err
	}

	objID2Path := make(map[string]string, len(rows))
	for _, row := range rows {
		objID2Path[row.DocObjectID] = row.Path
	}

	docHTML := gdocs.Convert(&dbDoc.Document)
	if n := xhtml.Select(docHTML, xhtml.WithAtom(atom.Data)); n != nil {
		return fmt.Errorf(
			"document unexpectedly contains <data> element: %q",
			xhtml.OuterHTML(n),
		)
	}

	// First remove everything after a ###
	removeTail(docHTML)

	// Now collect the embeds array and metadata
	var (
		metadata db.GDocsMetadata
		embeds   = []db.Embed{} // must not be "null"
		warnings = []string{}
		n        = 1
	)

	// Default slug is article title
	metadata.InternalID = dbDoc.Document.Title

	for tbl, rows := range xhtml.Tables(docHTML) {
		embed := db.Embed{N: n}
		switch label := rows.Label(); label {
		case "html", "embed", "raw", "script":
			embed.Type = db.RawEmbedTag
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			embed.Value = embedHTML
			if !ascii.Contains(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d contains unusual characters.", n,
				))
			}
			goto append

		case "spl", "spl-embed":
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			embed.Type = db.SpotlightRawEmbedOrTextTag
			embed.Value = embedHTML
			value := must.Get(json.Marshal(embed))
			data := xhtml.New("data", "value", string(value))
			xhtml.ReplaceWith(tbl, data)

		case "spl-text":
			embed.Type = db.SpotlightRawEmbedOrTextTag
			n := xhtml.Clone(rows.At(1, 0))
			blocko.MergeSiblings(n)
			blocko.RemoveEmptyP(n)
			blocko.RemoveMarks(n)
			s := blocko.Blockize(n)
			embed.Value = s
			value := must.Get(json.Marshal(embed))
			data := xhtml.New("data", "value", string(value))
			xhtml.ReplaceWith(tbl, data)

		case "partner-embed":
			embedHTML := xhtml.TextContent(rows.At(1, 0))
			embed.Type = db.PartnerRawEmbedTag
			embed.Value = embedHTML
			goto append

		case "partner-text":
			embed.Type = db.PartnerTextTag
			n := xhtml.Clone(rows.At(1, 0))
			blocko.MergeSiblings(n)
			blocko.RemoveEmptyP(n)
			blocko.RemoveMarks(n)
			embed.Value = xhtml.InnerHTMLBlocks(n)
			value := must.Get(json.Marshal(embed))
			data := xhtml.New("data", "value", string(value))
			xhtml.ReplaceWith(tbl, data)

		case "photo", "image", "photograph", "illustration", "illo":
			embed.Type = db.ImageEmbedTag
			if imageEmbed, warning := svc.replaceImageEmbed(
				ctx, tbl, rows, n, dbDoc.ExternalID, objID2Path,
			); warning != "" {
				tbl.Parent.RemoveChild(tbl)
				warnings = append(warnings, warning)
			} else {
				embed.Value = *imageEmbed
				goto append
			}

		case "metadata", "info":
			if warning := svc.replaceMetadata(
				ctx, tbl, rows, dbDoc.ExternalID, objID2Path, &metadata,
			); warning != "" {
				warnings = append(warnings, warning)
			}
			tbl.Parent.RemoveChild(tbl)

		case "comment", "ignore", "note":
			tbl.Parent.RemoveChild(tbl)

		case "table":
			row := xhtml.Closest(rows.At(0, 0), xhtml.WithAtom(atom.Tr))
			row.Parent.RemoveChild(row)

		case "toc", "table of contents":
			embed.Type = db.ToCEmbedTag
			embed.Value = processToc(docHTML, rows)
			goto append

		default:
			warnings = append(warnings, fmt.Sprintf(
				"Unrecognized table type: %q", label,
			))
			tbl.Parent.RemoveChild(tbl)
		}
		continue
	append:
		embeds = append(embeds, embed)
		value := must.Get(json.Marshal(embed))
		data := xhtml.New("data", "value", string(value))
		xhtml.ReplaceWith(tbl, data)
		n++
	}

	docHTML, err = blocko.Minify(xhtml.ToBuffer(docHTML))
	if err != nil {
		return err
	}
	blocko.MergeSiblings(docHTML)
	blocko.RemoveEmptyP(docHTML)
	blocko.RemoveMarks(docHTML)

	// Warn about fake headings
	for n := range xhtml.ChildNodes(docHTML) {
		// <p> with only b/i/strong/em for a child
		if n.DataAtom != atom.P {
			continue
		}
		if n.FirstChild != nil &&
			n.FirstChild == n.LastChild &&
			slices.Contains([]atom.Atom{
				atom.B, atom.Strong,
			}, n.FirstChild.DataAtom) {
			text := xhtml.TextContent(n)
			if len([]rune(text)) > 17 {
				runes := []rune(text)[:13]
				text = string(runes) + "..."
			}
			warning := fmt.Sprintf(
				"Paragraph beginning %q looks like a header, but does not use H-tag.", text)
			warnings = append(warnings, warning)
		}
	}

	// Warn about <br>
	if n := xhtml.Select(docHTML, xhtml.WithAtom(atom.Br)); n != nil {
		warnings = append(warnings,
			"Document contains <br> line breaks. Are you sure you want to use a line break? In Google Docs, select View > Show non-printing characters to see them.")
	}

	// Clone and remove turn data atoms into attributes
	richText := xhtml.Clone(docHTML)
	fixRichTextPlaceholders(richText)

	// For rawHTML, convert to raw nodes
	rawHTML := xhtml.Clone(docHTML)
	fixRawHTMLPlaceholders(rawHTML)

	// Markdown data conversion
	fixMarkdownPlaceholders(docHTML)
	md := blocko.Blockize(docHTML)

	// Save to database
	_, err = svc.Queries.UpdateGDocsDoc(ctx, db.UpdateGDocsDocParams{
		ID:              dbDoc.ID,
		Metadata:        metadata,
		Embeds:          embeds,
		RichText:        xhtml.InnerHTMLBlocks(richText),
		RawHtml:         xhtml.InnerHTMLBlocks(rawHTML),
		ArticleMarkdown: md,
		Warnings:        warnings,
		WordCount:       int32(stringx.WordCount(xhtml.TextContent(richText))),
	})
	return err
}

func removeTail(n *html.Node) {
	for c := range xhtml.ChildNodes(n) {
		if c.DataAtom == atom.Table {
			continue
		}

		if text := xhtml.TextContent(c); text != "###" {
			continue
		}

		remove := []*html.Node{c}
		for sibling := c.NextSibling; sibling != nil; sibling = sibling.NextSibling {
			remove = append(remove, sibling)
		}
		xhtml.RemoveAll(remove)
		return
	}
}

func replaceSpotlightShortcodes(s string) string {
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	// $("[data-spl-embed-version=1]")
	divs := xhtml.SelectSlice(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Div && xhtml.Attr(n, "data-spl-embed-version") == "1"
	})
	if len(divs) < 1 {
		return s
	}
	var buf strings.Builder
	for i, div := range divs {
		if i != 0 {
			buf.WriteString("\n")
		}
		netloc := xhtml.Attr(div, "data-spl-src")
		u, err := url.Parse(netloc)
		if err != nil {
			return s
		}
		tag := strings.Trim(u.Path, "/")
		if !slices.Contains([]string{
			"embeds/cta",
			"embeds/donate",
			"embeds/newsletter",
			"embeds/tips",
		}, tag) {
			return s
		}
		tag = strings.TrimPrefix(tag, "embeds/")
		q := u.Query()
		buf.WriteString("{{<embed/")
		buf.WriteString(tag)
		for _, k := range iterx.Sorted(iterx.Keys(q)) {
			vv := q[k]
			for _, v := range vv {
				buf.WriteString(" ")
				buf.WriteString(k)
				buf.WriteString("=\"")
				buf.WriteString(html.EscapeString(v))
				buf.WriteString("\"")
			}
		}
		buf.WriteString(">}}")
	}
	return buf.String()
}

func (svc Services) replaceImageEmbed(
	ctx context.Context,
	tbl *html.Node,
	rows xhtml.TableNodes,
	n int,
	externalID string,
	objID2Path map[string]string,
) (imageEmbed *db.EmbedImage, warning string) {
	var width, height int
	if w := xhtml.TextContent(rows.Value("width")); w != "" {
		width, _ = strconv.Atoi(w)
	}
	if h := xhtml.TextContent(rows.Value("height")); h != "" {
		height, _ = strconv.Atoi(h)
	}
	imageEmbed = &db.EmbedImage{
		Credit:  xhtml.TextContent(rows.Value("credit")),
		Caption: xhtml.TextContent(rows.Value("caption")),
		Description: cmp.Or(
			xhtml.TextContent(rows.Value("description")),
			xhtml.TextContent(rows.Value("alt")),
		),
		Width:  width,
		Height: height,
	}

	if path := xhtml.TextContent(rows.Value("path")); path != "" {
		imageEmbed.Path = path
		return imageEmbed, ""
	}

	linkTag := xhtml.Select(tbl, xhtml.WithAtom(atom.A))
	if href := xhtml.Attr(linkTag, "href"); href != "" {
		path, err := svc.ReplaceAndUploadImageURL(ctx, href, imageEmbed.Description, imageEmbed.Credit)
		switch {
		case err == nil:
			imageEmbed.Path = path
			return imageEmbed, ""
		case errors.Is(err, requests.ErrValidator):
			// Try looking up the image
		case err != nil:
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "ProcessGDocsDoc: ReplaceAndUploadImageURL", "err", err)
			return nil, fmt.Sprintf(
				"An error occurred when processing images in table %d: %v.",
				n, err)
		}
	}

	image := xhtml.Select(tbl, xhtml.WithAtom(atom.Img))
	if image == nil {
		return nil, fmt.Sprintf(
			"Table %d missing image", n,
		)
	}
	objID := xhtml.Attr(image, "data-oid")
	if path := objID2Path[objID]; path != "" {
		imageEmbed.Path = path
	} else {
		src := xhtml.Attr(image, "src")
		if uploadErr := svc.UploadGDocsImage(ctx, UploadGDocsImageParams{
			ExternalID:  externalID,
			DocObjectID: objID,
			ImageURL:    src,
			Embed:       imageEmbed,
		}); uploadErr != nil {
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "ProcessGDocsDoc: UploadGDocsImage", "err", uploadErr)
			return nil, fmt.Sprintf(
				"An error occurred when processing images in table %d: %v.",
				n, uploadErr)
		}
	}
	return imageEmbed, ""
}

func (svc Services) replaceMetadata(
	ctx context.Context,
	tbl *html.Node,
	rows xhtml.TableNodes,
	externalID string,
	objID2Path map[string]string,
	metadata *db.GDocsMetadata,
) string {
	metadata.InternalID = cmp.Or(
		xhtml.TextContent(rows.Value("slug")),
		xhtml.TextContent(rows.Value("internal id")),
		metadata.InternalID,
	)
	metadata.Byline = cmp.Or(
		xhtml.TextContent(rows.Value("byline")),
		xhtml.TextContent(rows.Value("authors")),
		xhtml.TextContent(rows.Value("author")),
		xhtml.TextContent(rows.Value("by")),
	)
	if strings.HasPrefix(metadata.Byline, "By ") ||
		strings.HasPrefix(metadata.Byline, "by ") {
		metadata.Byline = metadata.Byline[3:]
	}
	metadata.Budget = xhtml.TextContent(rows.Value("budget"))
	metadata.Hed = cmp.Or(
		xhtml.TextContent(rows.Value("hed")),
		xhtml.TextContent(rows.Value("title")),
		xhtml.TextContent(rows.Value("headline")),
		xhtml.TextContent(rows.Value("hedline")),
	)
	metadata.Description = cmp.Or(
		xhtml.TextContent(rows.Value("seo description")),
		xhtml.TextContent(rows.Value("description")),
		xhtml.TextContent(rows.Value("desc")),
	)
	metadata.LedeImageCredit = cmp.Or(
		xhtml.TextContent(rows.Value("lede image credit")),
		xhtml.TextContent(rows.Value("lead image credit")),
		xhtml.TextContent(rows.Value("credit")),
	)
	metadata.LedeImageCaption = cmp.Or(
		xhtml.TextContent(rows.Value("lede image caption")),
		xhtml.TextContent(rows.Value("lead image caption")),
		xhtml.TextContent(rows.Value("caption")),
	)
	metadata.LedeImageDescription = cmp.Or(
		xhtml.TextContent(rows.Value("lede image description")),
		xhtml.TextContent(rows.Value("lead image description")),
		xhtml.TextContent(rows.Value("lede image alt")),
		xhtml.TextContent(rows.Value("lead image alt")),
		xhtml.TextContent(rows.Value("alt")),
	)
	metadata.URLSlug = cmp.Or(
		xhtml.TextContent(rows.Value("url")),
		xhtml.TextContent(rows.Value("keywords")),
	)
	metadata.URLSlug = strings.TrimRight(metadata.URLSlug, "/")
	_, metadata.URLSlug, _ = stringx.LastCut(metadata.URLSlug, "/")
	metadata.URLSlug = stringx.SlugifyURL(metadata.URLSlug)

	metadata.Blurb = cmp.Or(
		xhtml.TextContent(rows.Value("blurb")),
		xhtml.TextContent(rows.Value("summary")),
	)
	metadata.LinkTitle = cmp.Or(
		xhtml.TextContent(rows.Value("link title")),
	)
	metadata.SEOTitle = cmp.Or(
		xhtml.TextContent(rows.Value("seo hed")),
		xhtml.TextContent(rows.Value("seo title")),
		xhtml.TextContent(rows.Value("seo headline")),
		xhtml.TextContent(rows.Value("seo hedline")),
	)
	metadata.OGTitle = cmp.Or(
		xhtml.TextContent(rows.Value("facebook hed")),
		xhtml.TextContent(rows.Value("facebook title")),
	)
	metadata.TwitterTitle = cmp.Or(
		xhtml.TextContent(rows.Value("twitter hed")),
		xhtml.TextContent(rows.Value("twitter title")),
	)
	metadata.Eyebrow = cmp.Or(
		xhtml.TextContent(rows.Value("eyebrow")),
		xhtml.TextContent(rows.Value("kicker")),
	)

	path := cmp.Or(
		xhtml.TextContent(rows.Value("lede image path")),
		xhtml.TextContent(rows.Value("lead image path")),
		xhtml.TextContent(rows.Value("path")),
	)
	if path != "" {
		metadata.LedeImage = path
		return ""
	}
	cell := rows.Value("lede image")
	if cell == nil {
		cell = rows.Value("lead image")
	}
	if cell == nil {
		return ""
	}

	linkTag := xhtml.Select(cell, xhtml.WithAtom(atom.A))
	if href := xhtml.Attr(linkTag, "href"); href != "" {
		path, err := svc.ReplaceAndUploadImageURL(ctx, href, metadata.LedeImageDescription, metadata.LedeImageCredit)
		switch {
		case err == nil:
			metadata.LedeImage = path
			return ""
		case errors.Is(err, requests.ErrValidator):
			// Try image URL next
		case err != nil:
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "ProcessGDocsDoc: replaceMetadata: ReplaceAndUploadImageURL",
				"err", err)
			return fmt.Sprintf("An error occurred when processing the lede image: %v.", err)
		}
	}

	image := xhtml.Select(tbl, xhtml.WithAtom(atom.Img))
	if image == nil {
		return ""
	}
	objID := xhtml.Attr(image, "data-oid")
	if path := objID2Path[objID]; path != "" {
		metadata.LedeImage = path
	} else {
		src := xhtml.Attr(image, "src")
		imageEmbed := db.EmbedImage{
			Credit:      metadata.LedeImageCredit,
			Caption:     metadata.LedeImageCaption,
			Description: metadata.LedeImageDescription,
		}
		if uploadErr := svc.UploadGDocsImage(ctx, UploadGDocsImageParams{
			ExternalID:  externalID,
			DocObjectID: objID,
			ImageURL:    src,
			Embed:       &imageEmbed,
		}); uploadErr != nil {
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "ProcessGDocsDoc: replaceMetadata: UploadGDocsImage",
				"err", uploadErr)
			return fmt.Sprintf("An error occurred when processing the lede image: %v.", uploadErr)
		}
		metadata.LedeImage = imageEmbed.Path
	}
	return ""
}

func fixRichTextPlaceholders(richText *html.Node) {
	embeds := xhtml.SelectSlice(richText, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		switch embed.Type {
		case db.SpotlightRawEmbedOrTextTag:
			dataEl.Parent.RemoveChild(dataEl)
			continue
		case db.PartnerTextTag:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value.(string),
			})
			continue
		case db.ImageEmbedTag, db.RawEmbedTag, db.ToCEmbedTag, db.PartnerRawEmbedTag:
			placeholder := xhtml.New("h2", "style", "color: red;")
			xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", embed.N))
			xhtml.ReplaceWith(dataEl, placeholder)
		default:
			panic("unknown embed type: " + embed.Type)
		}
	}
}

func extractEmbed(n *html.Node) db.Embed {
	var embed db.Embed
	must.Do(json.Unmarshal([]byte(xhtml.Attr(n, "value")), &embed))
	return embed
}

func fixRawHTMLPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.SelectSlice(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		switch embed.Type {
		case db.SpotlightRawEmbedOrTextTag:
			dataEl.Parent.RemoveChild(dataEl)
		case db.RawEmbedTag, db.ToCEmbedTag, db.PartnerRawEmbedTag, db.PartnerTextTag:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value.(string),
			})
		case db.ImageEmbedTag:
			placeholder := xhtml.New("h2", "style", "color: red;")
			xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", embed.N))
			xhtml.ReplaceWith(dataEl, placeholder)
		default:
			panic("unknown embed type: " + embed.Type)
		}
	}
}

func fixMarkdownPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.SelectSlice(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		switch embed.Type {
		case db.PartnerRawEmbedTag, db.PartnerTextTag:
			dataEl.Parent.RemoveChild(dataEl)
		case db.RawEmbedTag, db.SpotlightRawEmbedOrTextTag:
			data := replaceSpotlightShortcodes(embed.Value.(string))
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
		case db.ToCEmbedTag:
			container := xhtml.New("div")
			must.Do(xhtml.SetInnerHTML(container, embed.Value.(string)))
			xhtml.ReplaceWith(dataEl, container)
			xhtml.UnnestChildren(container)
		case db.ImageEmbedTag:
			image := embed.Value.(db.EmbedImage)
			var widthHeight string
			if image.Width != 0 {
				widthHeight = fmt.Sprintf(`width-ratio="%d" height-ratio="%d" `,
					image.Width, image.Height,
				)
			}
			data := fmt.Sprintf(
				`{{<picture src="%s" %sdescription="%s" caption="%s" credit="%s">}}`,
				image.Path,
				widthHeight,
				html.EscapeString(strings.TrimSpace(image.Description)),
				html.EscapeString(strings.TrimSpace(image.Caption)),
				html.EscapeString(strings.TrimSpace(image.Credit)),
			)
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
		default:
			panic("unknown embed type: " + embed.Type)
		}
	}
}

type UploadGDocsImageParams struct {
	ExternalID  string
	DocObjectID string
	ImageURL    string
	Embed       *db.EmbedImage // In-out param
}

func (svc Services) UploadGDocsImage(ctx context.Context, arg UploadGDocsImageParams) (err error) {
	defer errorx.Trace(&err)

	// Download the image + headers
	body, ct, err := FetchImageURL(ctx, svc.Client, arg.ImageURL)
	if err != nil {
		return err
	}

	// Hash the file
	hash := md5.Sum(body)

	// Look up file hash
	dbImage, err := svc.Queries.GetImageByMD5(ctx, hash[:])
	var imageID int64
	switch {
	// If it is found, return & save the relationship for next refresh
	case err == nil:
		arg.Embed.Path = dbImage.Path
		imageID = dbImage.ID

	// If it's not found, it needs to be uploaded & saved
	case db.IsNotFound(err):
		itype, err := imageTypeFromMIME(ct)
		if err != nil {
			return err
		}
		uploadPath := makeCASaddress(body, ct)
		arg.Embed.Path = uploadPath
		h := http.Header{"Content-Type": []string{ct}}
		if err := svc.ImageStore.WriteFile(ctx, uploadPath, h, body); err != nil {
			return err
		}
		record, err := svc.Queries.UpsertImageWithMD5(ctx, db.UpsertImageWithMD5Params{
			Path:        uploadPath,
			Type:        itype,
			Description: arg.Embed.Description,
			Credit:      arg.Embed.Credit,
			MD5:         hash[:],
			Bytes:       int64(len(hash)),
		})
		if err != nil {
			return err
		}
		imageID = record.ID
	// Other errors are bad
	case err != nil:
		return err
	}

	return svc.Queries.UpsertGDocsImage(ctx, db.UpsertGDocsImageParams{
		ExternalID:  arg.ExternalID,
		DocObjectID: arg.DocObjectID,
		ImageID:     imageID,
	})
}

func processToc(doc *html.Node, rows xhtml.TableNodes) string {
	type header struct {
		text  string
		id    string
		depth int
	}
	var headers []header
	for n := range xhtml.All(doc) {
		switch n.DataAtom {
		case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		default:
			continue
		}
		id := fmt.Sprintf("spl-heading-%d", len(headers)+1)
		xhtml.SetAttr(n, "id", id)
		depth := int(n.Data[1] - '0')
		headers = append(headers, header{xhtml.TextContent(n), id, depth})
	}
	container := xhtml.New("div")
	h3 := xhtml.New("h3")
	xhtml.AppendText(h3, cmp.Or(
		xhtml.TextContent(rows.At(0, 1)),
		xhtml.TextContent(rows.At(1, 0)),
		"Table of Contents",
	))
	container.AppendChild(h3)
	ul := xhtml.New("ul")
	container.AppendChild(ul)
	currentUl := ul
	lastDepth := 7 // Past H6, the maximum possible depth
	for _, h := range headers {
		// If this one is deeper or less deep than its predecessor,
		// add and remove ULs as needed
		d := h.depth
		for lastDepth > d {
			// If its out of order, just try to cope
			currentUl = cmp.Or(
				xhtml.Closest(currentUl.Parent, xhtml.WithAtom(atom.Ul)),
				currentUl,
			)
			d++
		}
		for lastDepth < d {
			newUl := xhtml.New("ul")
			lastLi := xhtml.LastChildOrNew(currentUl, "li")
			lastLi.AppendChild(newUl)
			currentUl = newUl
			d--
		}
		li := xhtml.New("li")
		p := xhtml.New("p")
		link := xhtml.New("a", "href", "#"+h.id)
		xhtml.AppendText(link, h.text)
		p.AppendChild(link)
		li.AppendChild(p)
		currentUl.AppendChild(li)
		lastDepth = h.depth
	}

	return xhtml.InnerHTML(container)
}

func makeCASaddress(body []byte, ct string) string {
	// https://en.wikipedia.org/wiki/Content-addressable_storage
	b := make([]byte, 0, crockford.LenMD5)
	b = crockford.AppendMD5(crockford.Lower, b, body)[:16]
	b = crockford.AppendPartition(b[:0], b, 4)
	ext, ok := strings.CutPrefix(ct, "image/")
	if !ok {
		ext = "bin"
	}
	return "cas/" + string(b) + "." + ext
}

type SharedArticleGDoc struct {
	*db.GDocsDoc
	Document string `json:"document,omitempty"`
}

type SharedArticle struct {
	*db.SharedArticle
	GDocs SharedArticleGDoc `json:"gdocs"`
}

func (svc Services) InflateSharedArticle(ctx context.Context, a *db.SharedArticle) (v any, err error) {
	defer errorx.Trace(&err)

	if a.SourceType != "gdocs" {
		return a, nil
	}
	var id int64
	if err = json.Unmarshal(a.RawData, &id); err != nil {
		return nil, err
	}

	doc, err := svc.Queries.GetGDocsByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return SharedArticle{a, SharedArticleGDoc{&doc, ""}}, err
}

func (svc Services) UpsertSharedArticleForGDoc(ctx context.Context, dbDoc *db.GDocsDoc, refreshMetadata bool) (*db.SharedArticle, error) {
	idJSON := must.Get(json.Marshal(dbDoc.ID))
	if refreshMetadata {
		art, err := svc.Queries.UpdateSharedArticleFromGDocs(ctx, db.UpdateSharedArticleFromGDocsParams{
			ExternalID:           dbDoc.ExternalID,
			RawData:              idJSON,
			InternalID:           dbDoc.Metadata.InternalID,
			Byline:               dbDoc.Metadata.Byline,
			Budget:               dbDoc.Metadata.Budget,
			Hed:                  dbDoc.Metadata.Hed,
			Description:          dbDoc.Metadata.Description,
			Blurb:                dbDoc.Metadata.Blurb,
			LedeImage:            dbDoc.Metadata.LedeImage,
			LedeImageCredit:      dbDoc.Metadata.LedeImageCredit,
			LedeImageDescription: dbDoc.Metadata.LedeImageDescription,
			LedeImageCaption:     dbDoc.Metadata.LedeImageCaption,
		})
		return &art, err
	}
	art, err := svc.Queries.UpsertSharedArticleFromGDocs(ctx, db.UpsertSharedArticleFromGDocsParams{
		ExternalID:           dbDoc.ExternalID,
		RawData:              idJSON,
		InternalID:           dbDoc.Metadata.InternalID,
		Byline:               dbDoc.Metadata.Byline,
		Budget:               dbDoc.Metadata.Budget,
		Hed:                  dbDoc.Metadata.Hed,
		Description:          dbDoc.Metadata.Description,
		Blurb:                dbDoc.Metadata.Blurb,
		LedeImage:            dbDoc.Metadata.LedeImage,
		LedeImageCredit:      dbDoc.Metadata.LedeImageCredit,
		LedeImageDescription: dbDoc.Metadata.LedeImageDescription,
		LedeImageCaption:     dbDoc.Metadata.LedeImageCaption,
	})
	return &art, err
}
