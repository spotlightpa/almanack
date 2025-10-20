package db_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
	docs "google.golang.org/api/docs/v1"
)

func TestProcessGDocsDoc(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	testfile.Run(t, "testdata/gdoc*", func(t *testing.T, path string) {
		t.Parallel()
		svc := almanack.Services{
			Queries:    q,
			Tx:         db.NewTxable(p),
			ImageStore: aws.NewBlobStore("mem://"),
			FileStore:  aws.NewBlobStore("mem://"),
			Gsvc:       new(google.Service),
			Client: &http.Client{
				Transport: reqtest.Replay(path),
			},
		}
		if os.Getenv("RECORD") != "" {
			svc.Client.Transport = reqtest.Caching(nil, path)
			cl, _ := svc.Gsvc.DriveClient(t.Context())
			cl.Transport = reqtest.Caching(cl.Transport, path)
			svc.Gsvc.SetMockClient(cl)
		} else {
			svc.Gsvc.SetMockClient(svc.Client)
		}

		var doc docs.Document
		testfile.ReadJSON(t, path+"/doc.json", &doc)
		// Run twice to test the already uploaded path
		for range 2 {
			dbDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
				ExternalID: fmt.Sprintf("abc123_%s", stringx.SlugifyURL(path)),
				Document:   doc,
			})
			be.NilErr(t, err)
			err = svc.ProcessGDocsDoc(ctx, dbDoc)
			be.NilErr(t, err)
			dbDoc, err = svc.Queries.GetGDocsByID(ctx, dbDoc.ID)
			be.NilErr(t, err)

			rt := be.Relaxed(t)

			testfile.Equal(rt, path+"/raw.html", dbDoc.RawHtml)
			testfile.Equal(rt, path+"/rich.html", dbDoc.RichText)
			testfile.Equal(rt, path+"/article.md", dbDoc.ArticleMarkdown)
			testfile.EqualJSON(rt, path+"/metadata.json", dbDoc.Metadata)
			testfile.EqualJSON(rt, path+"/warnings.json", dbDoc.Warnings)

			art, err := svc.UpsertSharedArticleForGDoc(ctx, &dbDoc, false)
			be.NilErr(t, err)
			date := time.Date(2020, time.March, 15, 20, 00, 00, 00, time.UTC)
			art.PublicationDate.Time = date
			swapInternalID := filepath.Base(path) // Set a unique slug
			art.InternalID, swapInternalID = swapInternalID, art.InternalID
			be.NilErr(t, svc.CreatePageFromGDocsDoc(ctx, art, "news"))
			be.True(t, art.PageID.Valid)
			page, err := svc.Queries.GetPageByID(ctx, art.PageID.Int64)
			be.NilErr(t, err)
			// Swap internal ID back
			art.InternalID = swapInternalID
			// Stablize racey fields
			art.ID = 123
			art.PageID.Valid = false
			art.RawData = nil
			art.CreatedAt = date
			art.UpdatedAt = date
			testfile.EqualJSON(rt, path+"/shared-article.json", art)
			page.ID = 123
			page.CreatedAt = date
			page.UpdatedAt = date
			testfile.EqualJSON(rt, path+"/page.json", page)
		}
	})
}
