package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
	docHTML, err = blocko.Minify(xhtml.ToBuffer(docHTML))
	if err != nil {
		return err
	}
	blocko.MergeSiblings(docHTML)
	blocko.RemoveEmptyP(docHTML)

	if n := xhtml.Find(docHTML, xhtml.WithAtom(atom.Data)); n != nil {
		return fmt.Errorf(
			"document unexpectedly contains <data> element: %q",
			xhtml.ToString(n),
		)
	}

	// First collect the embeds array
	var embeds []db.Embed
	var warnings []string
	n := 1

	xhtml.Tables(docHTML, func(tbl *html.Node, rows xhtml.TableNodes) {
		label := rows.Label()
		embed := db.Embed{N: n}
		if slices.Contains([]string{"html", "embed", "raw", "script"}, label) {
			embed.Type = db.RawEmbedTag
			embedHTML := xhtml.InnerText(rows.At(1, 0))
			embed.Value = embedHTML
			if nonASCII.Contains(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d contains unusual characters.", n,
				))
			}
		} else if slices.Contains([]string{
			"photo", "image", "photograph", "illustration", "illo",
		}, label) {
			embed.Type = db.ImageEmbedTag
			if imageEmbed, warning := svc.replaceImageEmbed(
				ctx, tbl, rows, n, dbDoc.ExternalID, objID2Path,
			); warning != "" {
				tbl.Parent.RemoveChild(tbl)
				warnings = append(warnings, warning)
				return
			} else {
				embed.Value = *imageEmbed
			}
		} else if label == "metadata" {
			tbl.Parent.RemoveChild(tbl)
			return
		} else {
			warnings = append(warnings, fmt.Sprintf(
				"Unrecognized table type: %q", label,
			))
			tbl.Parent.RemoveChild(tbl)
			return
		}
		embeds = append(embeds, embed)
		value := must.Get(json.Marshal(embed))
		data := xhtml.New("data", "value", string(value))
		xhtml.ReplaceWith(tbl, data)
		n++
	})

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
		Embeds:          embeds,
		RichText:        xhtml.InnerBlocksToString(richText),
		RawHtml:         xhtml.InnerBlocksToString(rawHTML),
		ArticleMarkdown: md,
		Warnings:        warnings,
		WordCount:       int32(stringx.WordCount(xhtml.InnerText(richText))),
	})
	return err
}

func (svc Services) replaceImageEmbed(
	ctx context.Context,
	tbl *html.Node,
	rows xhtml.TableNodes,
	n int,
	externalID string,
	objID2Path map[string]string,
) (imageEmbed *db.EmbedImage, warning string) {
	imageEmbed = &db.EmbedImage{
		Credit:  xhtml.InnerText(rows.Value("credit")),
		Caption: xhtml.InnerText(rows.Value("caption")),
		Description: stringx.First(
			xhtml.InnerText(rows.Value("description")),
			xhtml.InnerText(rows.Value("alt")),
		),
	}

	if path := xhtml.InnerText(rows.Value("path")); path != "" {
		imageEmbed.Path = path
		return imageEmbed, ""
	}

	linkTag := xhtml.Find(tbl, xhtml.WithAtom(atom.A))
	if href := xhtml.Attr(linkTag, "href"); href != "" {
		path, err := svc.ReplaceAndUploadImageURL(ctx, href, imageEmbed.Description, imageEmbed.Credit)
		if err != nil {
			l := almlog.FromContext(ctx)
			l.ErrorCtx(ctx, "ProcessGDocsDoc: ReplaceAndUploadImageURL", "err", err)
			return nil, fmt.Sprintf(
				"An error occurred when processing images in table %d: %v.",
				n, err)
		}
		imageEmbed.Path = path
		return imageEmbed, ""
	}

	image := xhtml.Find(tbl, xhtml.WithAtom(atom.Img))
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
			l.ErrorCtx(ctx, "ProcessGDocsDoc: UploadGDocsImage", "err", uploadErr)
			return nil, fmt.Sprintf(
				"An error occurred when processing images in table %d: %v.",
				n, uploadErr)
		}
	}
	return imageEmbed, ""
}

func fixRichTextPlaceholders(richText *html.Node) {
	embeds := xhtml.FindAll(richText, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		placeholder := xhtml.New("h2", "style", "color: red;")
		xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", embed.N))
		xhtml.ReplaceWith(dataEl, placeholder)
	}
}

func extractEmbed(n *html.Node) db.Embed {
	var embed db.Embed
	must.Do(json.Unmarshal([]byte(xhtml.Attr(n, "value")), &embed))
	return embed
}

func fixRawHTMLPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.FindAll(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		switch embed.Type {
		case db.RawEmbedTag:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value.(string),
			})
		case db.ImageEmbedTag:
			placeholder := xhtml.New("h2", "style", "color: red;")
			xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", embed.N))
			xhtml.ReplaceWith(dataEl, placeholder)
		}
	}
}

func fixMarkdownPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.FindAll(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractEmbed(dataEl)
		switch embed.Type {
		case db.RawEmbedTag:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value.(string),
			})
		case db.ImageEmbedTag:
			// TODO: distinguish image types
			image := embed.Value.(db.EmbedImage)
			data := fmt.Sprintf(
				`{{<picture src="%s" description="%s" caption="%s" credit="%s">}}`,
				image.Path,
				strings.TrimSpace(image.Description),
				strings.TrimSpace(image.Caption),
				strings.TrimSpace(image.Credit),
			)
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
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
	uploadPath := makeCASaddress(body, ct)

	// Look up file hash
	dbImage, err := svc.Queries.GetImageByPath(ctx, uploadPath)
	var imageID int64
	switch {
	// If it's not found, it needs to be uploaded & saved
	case db.IsNotFound(err):
		record, err := svc.uploadAndRecordImage(ctx, uploadAndRecordImageParams{
			UploadPath:  uploadPath,
			Body:        body,
			ContentType: ct,
			Description: arg.Embed.Description,
			Credit:      arg.Embed.Credit,
			SourceURL:   "",
		})
		if err != nil {
			return err
		}
		imageID = record.ID
	// Other errors are bad
	case err != nil:
		return err
	// If it is found, return & save the relationship for next refresh
	case err == nil:
		imageID = dbImage.ID
	}

	arg.Embed.Path = uploadPath
	return svc.Queries.UpsertGDocsImage(ctx, db.UpsertGDocsImageParams{
		ExternalID:  arg.ExternalID,
		DocObjectID: arg.DocObjectID,
		ImageID:     imageID,
	})
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

var nonASCII = bytemap.Range(128, 255)

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
