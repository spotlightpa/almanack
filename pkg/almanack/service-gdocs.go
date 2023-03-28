package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/slicex"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (svc Services) SharedArticleFromGDocs(ctx context.Context, id string) (obj any, err error) {
	defer errorx.Trace(&err)

	if err := svc.ConfigureGoogleCert(ctx); err != nil {
		return nil, err
	}
	cl, err := svc.Gsvc.GdocsClient(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := gdocs.Request(ctx, cl, id)
	if err != nil {
		// TODO: figure out common errors, like no-permissions
		return nil, err
	}

	// TODO: Extract metadata

	dbDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
		GDocsID:  id,
		Document: *doc,
	})
	if err != nil {
		return nil, err
	}

	idJSON, err := json.Marshal(dbDoc.ID)
	if err != nil {
		return nil, err
	}

	return svc.Queries.UpsertSharedArticleFromGDocs(ctx, db.UpsertSharedArticleFromGDocsParams{
		GdocsID:    id,
		InternalID: doc.Title,
		RawData:    idJSON,
	})
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
	rows, err := svc.Queries.ListGDocsImagesByGDocsID(ctx, dbDoc.GDocsID)
	if err != nil {
		return err
	}

	objID2Path := make(map[string]string, len(rows))
	for _, row := range rows {
		objID2Path[row.DocObjectID] = row.Path
	}

	rawHTML := gdocs.Convert(&dbDoc.Document)
	blocko.MergeSiblings(rawHTML)
	blocko.RemoveEmptyP(rawHTML)

	// First collect the embeds array
	var embeds []db.Embed
	var warnings []string
	n := 0
	var imageUploadErrs []error

	xhtml.Tables(rawHTML, func(tbl *html.Node, rows xhtml.TableNodes) {
		label := rows.Label()
		embedHTML := ""
		if slices.Contains([]string{"html", "embed", "raw", "script"}, label) {
			embedHTML = xhtml.InnerText(rows.At(1, 0))
			embedHTML = strings.TrimSpace(embedHTML)
			embeds = append(embeds, db.Embed{
				N:     n,
				Type:  "raw",
				Value: embedHTML,
			})
			if nonASCII.Contains(embedHTML) {
				warnings = append(warnings, fmt.Sprintf(
					"Embed #%d contains unusual characters.", n,
				))
			}
		} else if slices.Contains([]string{
			"photo", "image", "photograph", "illustration", "illo",
		}, label) {
			imageEmbed := db.EmbedImage{
				Credit:  rows.Value("credit"),
				Caption: rows.Value("caption"),
				Description: stringx.First(
					rows.Value("alt"),
					rows.Value("description")),
			}
			image := xhtml.Find(tbl, xhtml.WithAtom(atom.Img))
			objID := xhtml.Attr(image, "data-oid")
			path := objID2Path[objID]
			if path == "" {
				src := xhtml.Attr(image, "src")
				if uploadErr := svc.UploadGDocsImage(ctx, UploadGDocsImageParams{
					GDocsID:     dbDoc.GDocsID,
					DocObjectID: objID,
					ImageURL:    src,
					Embed:       &imageEmbed,
				}); uploadErr != nil {
					imageUploadErrs = append(imageUploadErrs, uploadErr)
					return
				}
			}

			embeds = append(embeds, db.Embed{
				N:     n + 1,
				Type:  "image",
				Value: imageEmbed,
			})
		} else {
			warnings = append(warnings, fmt.Sprintf(
				"Unrecognized table type: %q", label,
			))
			tbl.Parent.RemoveChild(tbl)
			return
		}
		n++
		placeholder := xhtml.New("h2", "style", "color: red;")
		xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", n))
		if embedHTML != "" {
			placeholder.Attr = append(placeholder.Attr, html.Attribute{
				Key: "raw_html",
				Val: embedHTML,
			})
		}
		tbl.Parent.InsertBefore(placeholder, tbl)
		tbl.Parent.RemoveChild(tbl)
	})

	if len(imageUploadErrs) > 0 {
		return errors.Join(imageUploadErrs...)
	}

	// Clone and remove raw_html attributes
	richText := xhtml.Clone(rawHTML)
	xhtml.VisitAll(richText, func(n *html.Node) {
		if n.DataAtom == atom.H2 && xhtml.Attr(n, "raw_html") != "" {
			slicex.DeleteFunc(&n.Attr, func(a html.Attribute) bool {
				return a.Key == "raw_html"
			})
		}
	})

	// For rawHTML, convert to raw nodes
	htmlEmbeds := xhtml.FindAll(rawHTML, func(n *html.Node) bool {
		return n.DataAtom == atom.H2 && xhtml.Attr(n, "raw_html") != ""
	})
	for _, n := range htmlEmbeds {
		rawNode := &html.Node{
			Type: html.RawNode,
			Data: xhtml.Attr(n, "raw_html"),
		}
		n.Parent.InsertBefore(rawNode, n)
		n.Parent.RemoveChild(n)
	}

	// TODO: Markdown conversion
	mdDoc := xhtml.Clone(rawHTML)
	md, err := blocko.HTMLToMarkdown(xhtml.ContentsToString(mdDoc))
	if err != nil {
		return err
	}

	// Save to database
	_, err = svc.Queries.UpdateGDocsDoc(ctx, db.UpdateGDocsDocParams{
		ID:              dbDoc.ID,
		Embeds:          embeds,
		RichText:        xhtml.ContentsToString(richText),
		RawHtml:         xhtml.ContentsToString(rawHTML),
		ArticleMarkdown: md,
		Warnings:        warnings,
		WordCount:       int32(stringx.WordCount(xhtml.InnerText(richText))),
	})
	return err
}

type UploadGDocsImageParams struct {
	GDocsID     string
	DocObjectID string
	ImageURL    string
	Embed       *db.EmbedImage // In-out param
}

func (svc Services) UploadGDocsImage(ctx context.Context, arg UploadGDocsImageParams) (err error) {
	// Download the image + headers
	body, ct, err := FetchImageURL(ctx, svc.Client, arg.ImageURL)
	if err != nil {
		return err
	}

	itype, ok := strings.CutPrefix(ct, "image/")
	if !ok {
		return fmt.Errorf("bad image content-type: %q", ct)
	}

	// Hash the file
	uploadPath := makeCASaddress(body, ct)

	// Look up file hash
	dbImage, err := svc.Queries.GetImageByPath(ctx, uploadPath)
	switch {
	// If it's not found, it needs to be uploaded & saved
	case db.IsNotFound(err):
		// Upload file
		h := make(http.Header, 1)
		h.Set("Content-Type", ct)
		if err = svc.ImageStore.WriteFile(ctx, uploadPath, h, body); err != nil {
			return err
		}

		// Save file hash as an image
		dbImage, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
			Path:        uploadPath,
			Type:        itype,
			Description: arg.Embed.Description,
			Credit:      arg.Embed.Credit,
			IsUploaded:  true,
		})
		if err != nil {
			return err
		}
	// Other errors are bad
	case err != nil:
		return err
	// If it is found, return & save the relationship for next refresh
	case err == nil:
		break
	}

	arg.Embed.Path = dbImage.Path
	err = svc.Queries.UpsertGDocsImage(ctx, db.UpsertGDocsImageParams{
		GDocsID:     arg.GDocsID,
		DocObjectID: arg.DocObjectID,
		ImageID:     dbImage.ID,
	})
	return err
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

type SharedArticle struct {
	*db.SharedArticle
	GDocs *db.GDocsDoc `json:"gdocs"`
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

	return SharedArticle{a, &doc}, err
}
