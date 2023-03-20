package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/workgroup"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"google.golang.org/api/docs/v1"
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

	// Extract images
	n := gdocs.Convert(doc)
	if err = svc.EnsureImages(ctx, id, n); err != nil {
		return nil, err
	}

	b, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	return svc.Queries.UpsertSharedArticleFromGDocs(ctx, db.UpsertSharedArticleFromGDocsParams{
		GdocsID:    id,
		InternalID: doc.Title,
		RawData:    b,
	})
}

func (svc Services) EnsureImages(ctx context.Context, id string, n *html.Node) (err error) {
	imgs := xhtml.FindAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Img
	})
	if len(imgs) < 1 {
		return nil
	}
	var pairs [][2]string
	for _, img := range imgs {
		imageID := xhtml.Attr(img, "data-oid")
		srcURL := xhtml.Attr(img, "src")
		pairs = append(pairs, [2]string{imageID, srcURL})
	}
	pairsBytes, err := json.Marshal(pairs)
	if err != nil {
		return err
	}
	return svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) error {
		if err = q.DeleteGDocsImagesWhereUnset(ctx, id); err != nil {
			return err
		}

		return q.UpsertGDocsIDObjectID(ctx, db.UpsertGDocsIDObjectIDParams{
			GDocsID:        id,
			ObjectUrlPairs: pairsBytes,
		})
	})
}

func (svc Services) UploadGoogleImages(ctx context.Context) (err error) {
	images, err := svc.Queries.ListGDocsImagesWhereUnset(ctx)
	if err != nil {
		return err
	}

	return workgroup.DoTasks(5, images, func(image db.GDocsImage) error {
		return svc.UploadGoogleImage(ctx, image)
	})
}

func (svc Services) UploadGoogleImage(ctx context.Context, gdi db.GDocsImage) (err error) {
	defer errorx.Trace(&err)

	// Download the image + headers
	body, ct, err := FetchImageURL(ctx, svc.Client, gdi.SourceURL)
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
	case err == nil:
		_, err = svc.Queries.UpdateGDocsImage(ctx, db.UpdateGDocsImageParams{
			ID:      gdi.ID,
			ImageID: pgtype.Int8{Int64: dbImage.ID, Valid: true},
		})
		return err
	case db.IsNotFound(err):
		break
	case err != nil:
		return err
	}
	// Upload file
	h := make(http.Header, 1)
	h.Set("Content-Type", ct)
	if err = svc.ImageStore.WriteFile(ctx, uploadPath, h, body); err != nil {
		return err
	}

	// Save file hash as an image
	dbImage, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:       uploadPath,
		Type:       itype,
		IsUploaded: true,
	})
	if err != nil {
		return err
	}

	// Update the google image table
	_, err = svc.Queries.UpdateGDocsImage(ctx, db.UpdateGDocsImageParams{
		ID:      gdi.ID,
		ImageID: pgtype.Int8{Int64: dbImage.ID, Valid: true},
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
	return fmt.Sprintf("cas/%s.%s", b, ext)
}

func (svc Services) InflateSharedArticle(ctx context.Context, a *db.SharedArticle) (v any, err error) {
	defer errorx.Trace(&err)

	if a.SourceType != "gdocs" {
		return a, nil
	}
	rows, err := svc.Queries.ListGDocsImagesByGDocsID(ctx, a.SourceID)
	if err != nil {
		return nil, err
	}
	// Warn if it has outstanding images
	for _, row := range rows {
		if !row.IsUploaded.Bool {
			type response struct {
				*db.SharedArticle
				RawData      string `json:"raw_data"`
				IsProcessing bool   `json:"is_processing"`
			}
			return response{a, "", true}, nil
		}
	}

	var doc docs.Document
	if err = json.Unmarshal(a.RawData, &doc); err != nil {
		return nil, err
	}
	root := gdocs.Convert(&doc)
	var items []string
	for el := root.FirstChild; el != nil; el = el.NextSibling {
		items = append(items, xhtml.ToString(el))
	}
	type response struct {
		*db.SharedArticle
		RawData string   `json:"raw_data"`
		Items   []string `json:"items"`
	}
	return response{a, "", items}, nil
}
