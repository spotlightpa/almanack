package almanack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/workgroup"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
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

	// Extract images
	n := gdocs.Convert(doc)
	if err = svc.EnsureImages(id, n); err != nil {
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

type imageData struct {
	URL, DocID, ImageID, Description, Credit, Caption string
}

func (svc Services) EnsureImages(id string, n *html.Node) (err error) {
	var imgs []imageData
	xhtml.Tables(n, func(tbl *html.Node, rows xhtml.TableNodes) {
		if !slices.Contains([]string{
			"picture",
			"photo",
			"photograph",
			"graphic",
			"illustration",
			"illo",
		}, rows.Label()) {
			return
		}
		img := imageData{DocID: id}
		xhtml.BreadFirst(tbl, func(n *html.Node) bool {
			if n.DataAtom == atom.Img {
				img.URL = xhtml.Attr(n, "src")
				img.ImageID = xhtml.Attr(n, "data-oid")
				return xhtml.Done
			}
			return xhtml.Continue
		})
		img.Credit = rows.Value("credit")
		img.Description = stringx.First(
			rows.Value("description"),
			rows.Value("alt"))
		img.Caption = rows.Value("caption")
		imgs = append(imgs, img)
	})

	return workgroup.DoTasks(workgroup.MaxProcs, imgs, svc.saveImage)
}

func (svc Services) saveImage(img imageData) (err error) {
	ctx := context.Background()
	itype, err := svc.typeForImage(ctx, img.URL)
	if err != nil {
		return err
	}
	path := gdocsHashPath(img.DocID, img.ImageID, itype)
	_, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:        path,
		Type:        itype,
		Description: img.Description,
		Credit:      img.Credit,
		SourceURL:   img.URL,
		IsUploaded:  false,
	})
	return err
}

func gdocsHashPath(docID, imageID, ext string) string {
	prefix := fmt.Append(nil, "spl:", docID)
	prefixHash := crockford.MD5(crockford.Lower, prefix)[:16]
	prefixHash = crockford.Partition(prefixHash, 4)
	suffix := fmt.Append(nil, "spl:", imageID)
	suffixHash := crockford.MD5(crockford.Lower, suffix)[:12]
	suffixHash = crockford.Partition(suffixHash, 4)
	return fmt.Sprintf("docs/%s/%s.%s", prefixHash, suffixHash, ext)
}
