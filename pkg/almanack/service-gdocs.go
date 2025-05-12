package almanack

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/tableaux"
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

	var warnings []string

	// Handle image uploads/database lookups
	for tbl, rows := range tableaux.Tables(docHTML) {
		switch label := rows.Label(); label {
		case "photo", "image", "photograph", "illustration", "illo",
			"spl-photo", "partner-photo", "spl-image", "partner-image":
			if warning := svc.replaceImagePath(
				ctx, tbl, rows, dbDoc.ExternalID, objID2Path,
			); warning != "" {
				warnings = append(warnings, warning)
			}

		case "metadata", "info":
			if warning := svc.replaceMetadataImagePath(
				ctx, tbl, rows, dbDoc.ExternalID, objID2Path,
			); warning != "" {
				warnings = append(warnings, warning)
			}
		}
	}

	metadata, embeds, richText, rawHTML, md, warnings2 := processDocHTML(docHTML)
	warnings = append(warnings, warnings2...)

	// Default slug is article title
	metadata.InternalID = cmp.Or(metadata.InternalID, dbDoc.Document.Title)

	// database slices must not be "null"
	if embeds == nil {
		embeds = []db.Embed{}
	}
	if warnings == nil {
		warnings = []string{}
	}

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
	for c := range n.ChildNodes() {
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

func (svc Services) replaceImagePath(
	ctx context.Context,
	tbl *html.Node,
	rows tableaux.TableNodes,
	externalID string,
	objID2Path map[string]string,
) (warning string) {
	if path := xhtml.TextContent(rows.Value("path")); path != "" {
		return ""
	}

	imageEmbed := &db.EmbedImage{
		Credit:  xhtml.TextContent(rows.Value("credit")),
		Caption: xhtml.TextContent(rows.Value("caption")),
		Description: cmp.Or(
			xhtml.TextContent(rows.Value("description")),
			xhtml.TextContent(rows.Value("alt")),
		),
	}

	linkTag := xhtml.Select(tbl, xhtml.WithAtom(atom.A))
	if href := xhtml.Attr(linkTag, "href"); href != "" {
		path, err := svc.ReplaceAndUploadImageURL(ctx, href, imageEmbed.Description, imageEmbed.Credit)
		switch {
		case err == nil:
			setRowValue(tbl, "path", path)
			return ""

		case errors.Is(err, requests.ErrValidator):
			// Try looking up the image
			break
		case err != nil:
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "ProcessGDocsDoc: ReplaceAndUploadImageURL", "err", err)
			return fmt.Sprintf(
				"An error occurred when processing images in table: %v.", err)
		}
	}

	image := xhtml.Select(tbl, xhtml.WithAtom(atom.Img))
	if image == nil {
		return ""
	}
	objID := xhtml.Attr(image, "data-oid")
	if path := objID2Path[objID]; path != "" {
		setRowValue(tbl, "path", path)
		return ""
	}
	src := xhtml.Attr(image, "src")
	if uploadErr := svc.UploadGDocsImage(ctx, UploadGDocsImageParams{
		ExternalID:  externalID,
		DocObjectID: objID,
		ImageURL:    src,
		Embed:       imageEmbed,
	}); uploadErr != nil {
		l := almlog.FromContext(ctx)
		l.ErrorContext(ctx, "ProcessGDocsDoc: UploadGDocsImage", "err", uploadErr)
		return fmt.Sprintf(
			"An error occurred when processing images in table: %v.", uploadErr)
	}

	setRowValue(tbl, "path", imageEmbed.Path)
	return ""
}

func setRowValue(tbl *html.Node, key, value string) {
	tr := xhtml.New("tr")
	keyNode := xhtml.New("td")
	xhtml.AppendText(keyNode, key)
	tr.AppendChild(keyNode)

	valueNode := xhtml.New("td")
	xhtml.AppendText(valueNode, value)
	tr.AppendChild(valueNode)

	tbl.AppendChild(tr)
}

func (svc Services) replaceMetadataImagePath(
	ctx context.Context,
	tbl *html.Node,
	rows tableaux.TableNodes,
	externalID string,
	objID2Path map[string]string,
) string {
	if path := cmp.Or(
		xhtml.TextContent(rows.Value("lede image path")),
		xhtml.TextContent(rows.Value("lead image path")),
		xhtml.TextContent(rows.Value("path")),
	); path != "" {
		return ""
	}
	cell := rows.Value("lede image")
	if cell == nil {
		cell = rows.Value("lead image")
	}
	if cell == nil {
		return ""
	}
	credit := cmp.Or(
		xhtml.TextContent(rows.Value("lede image credit")),
		xhtml.TextContent(rows.Value("lead image credit")),
		xhtml.TextContent(rows.Value("credit")),
	)
	description := cmp.Or(
		xhtml.TextContent(rows.Value("lede image description")),
		xhtml.TextContent(rows.Value("lead image description")),
		xhtml.TextContent(rows.Value("lede image alt")),
		xhtml.TextContent(rows.Value("lead image alt")),
		xhtml.TextContent(rows.Value("alt")),
	)

	linkTag := xhtml.Select(cell, xhtml.WithAtom(atom.A))
	if href := xhtml.Attr(linkTag, "href"); href != "" {
		path, err := svc.ReplaceAndUploadImageURL(ctx, href, description, credit)
		switch {
		case err == nil:
			setRowValue(tbl, "path", path)
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
		setRowValue(tbl, "path", path)
		return ""
	}

	src := xhtml.Attr(image, "src")
	imageEmbed := db.EmbedImage{
		Credit:      credit,
		Description: description,
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
	setRowValue(tbl, "path", imageEmbed.Path)
	return ""
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
