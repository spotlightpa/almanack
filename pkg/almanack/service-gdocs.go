package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
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
	cl, err := svc.Gsvc.GdocsClient(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := gdocs.Request(ctx, cl, externalID)
	if err != nil {
		// TODO: figure out common errors, like no-permissions
		return nil, err
	}

	// TODO: Extract metadata

	newDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
		GDocsID:  externalID,
		Document: *doc,
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
	rows, err := svc.Queries.ListGDocsImagesByGDocsID(ctx, dbDoc.GDocsID)
	if err != nil {
		return err
	}

	objID2Path := make(map[string]string, len(rows))
	for _, row := range rows {
		objID2Path[row.DocObjectID] = row.Path
	}

	docHTML := gdocs.Convert(&dbDoc.Document)
	docHTML, err = blocko.Minify(xhtml.ToString(docHTML))
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
			embed.Type = "raw"
			embedHTML := xhtml.InnerText(rows.At(1, 0))
			embedHTML = strings.TrimSpace(embedHTML)
			embed.Value = embedHTML
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
			// TODO: If there's a link, use that instead
			image := xhtml.Find(tbl, xhtml.WithAtom(atom.Img))
			if image == nil {
				warnings = append(warnings, fmt.Sprintf(
					"Table %d missing image", n,
				))
				tbl.Parent.RemoveChild(tbl)
				return
			}
			objID := xhtml.Attr(image, "data-oid")
			if path := objID2Path[objID]; path != "" {
				imageEmbed.Path = path
			} else {
				src := xhtml.Attr(image, "src")
				if uploadErr := svc.UploadGDocsImage(ctx, UploadGDocsImageParams{
					GDocsID:     dbDoc.GDocsID,
					DocObjectID: objID,
					ImageURL:    src,
					Embed:       &imageEmbed,
				}); uploadErr != nil {
					l := almlog.FromContext(ctx)
					l.ErrorCtx(ctx, "ProcessGDocsDoc: UploadGDocsImage", "err", uploadErr)
					warnings = append(warnings, fmt.Sprintf(
						"An error occurred when processing images in table %d: %v.",
						n, uploadErr))
					tbl.Parent.RemoveChild(tbl)
					return
				}
			}
			embed.Type = "image"
			embed.Value = imageEmbed
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
		value, err := json.Marshal(embed.Value)
		if err != nil {
			panic(err)
		}
		data := xhtml.New("data",
			"n", strconv.Itoa(embed.N),
			"type", embed.Type,
			"value", string(value))

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

func fixRichTextPlaceholders(richText *html.Node) {
	embeds := xhtml.FindAll(richText, xhtml.WithAtom(atom.Data))
	for _, embed := range embeds {
		placeholder := xhtml.New("h2", "style", "color: red;")
		n := xhtml.Attr(embed, "n")
		xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%s", n))
		xhtml.ReplaceWith(embed, placeholder)
	}
}

func fixRawHTMLPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.FindAll(rawHTML, xhtml.WithAtom(atom.Data))
	for _, embed := range embeds {
		dataType := xhtml.Attr(embed, "type")
		value := xhtml.Attr(embed, "value")
		if dataType == "raw" {
			var data string
			if err := json.Unmarshal([]byte(value), &data); err != nil {
				panic(err)
			}
			xhtml.ReplaceWith(embed, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
			continue
		}
		placeholder := xhtml.New("h2", "style", "color: red;")
		n := xhtml.Attr(embed, "n")
		xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%s", n))
		xhtml.ReplaceWith(embed, placeholder)
	}
}

func fixMarkdownPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.FindAll(rawHTML, xhtml.WithAtom(atom.Data))
	for _, embed := range embeds {
		dataType := xhtml.Attr(embed, "type")
		value := xhtml.Attr(embed, "value")
		if dataType == "raw" {
			var data string
			if err := json.Unmarshal([]byte(value), &data); err != nil {
				panic(err)
			}
			xhtml.ReplaceWith(embed, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
			continue
		}
		var image db.EmbedImage
		if err := json.Unmarshal([]byte(value), &image); err != nil {
			panic(err)
		}

		// TODO: distinguish image types
		data := fmt.Sprintf(
			`{{<picture src="%s" description="%s" caption="%s" credit="%s">}}`,
			image.Path,
			strings.TrimSpace(image.Description),
			strings.TrimSpace(image.Caption),
			strings.TrimSpace(image.Credit),
		)
		xhtml.ReplaceWith(embed, &html.Node{
			Type: html.RawNode,
			Data: data,
		})
	}
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
