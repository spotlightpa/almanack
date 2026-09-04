package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/google"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"github.com/spotlightpa/almanack/internal/utils/timex"
	docs "google.golang.org/api/docs/v1"
)

func TestProcessGDocsDoc(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)

	ctx := t.Context()
	testfile.Run(t, "testdata/gdoc*", func(t *testing.T, path string) {
		be := assert.FailNow(t)
		t.Parallel()
		svc := almsvc.Services{
			DB:         dbhandle,
			Queries:    dbhandle.Queries(),
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
			dbDoc := be.OK(svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
				ExternalID: fmt.Sprintf("abc123_%s", stringx.SlugifyURL(path)),
				Document:   doc,
			}))
			be.Zero(svc.ProcessGDocsDoc(ctx, dbDoc))
			dbDoc, err := svc.Queries.GetGDocsByID(ctx, dbDoc.ID)
			be.Zero(err)

			testfile.Equal(t, path+"/raw.html", dbDoc.RawHtml)
			testfile.Equal(t, path+"/rich.html", dbDoc.RichText)
			testfile.Equal(t, path+"/article.md", dbDoc.ArticleMarkdown)
			testfile.EqualJSON(t, path+"/metadata.json", dbDoc.Metadata)
			testfile.EqualJSON(t, path+"/warnings.json", dbDoc.Warnings)

			art := be.OK(svc.UpsertSharedArticleForGDoc(ctx, &dbDoc, false))
			date := time.Date(2020, time.March, 15, 20, 00, 00, 00, time.UTC)
			art.PublicationDate.Time = date
			swapInternalID := filepath.Base(path) // Set a unique slug
			art.InternalID, swapInternalID = swapInternalID, art.InternalID
			be.
				Zero(svc.CreatePageFromGDocsDoc(ctx, art, "news")).
				True(art.PageID.Valid)
			page := be.OK(svc.Queries.GetPageByID(ctx, art.PageID.Int64))
			// Swap internal ID back
			art.InternalID = swapInternalID
			// Stablize racey fields
			art.ID = 123
			art.PageID.Valid = false
			art.RawData = nil
			art.CreatedAt = date
			art.UpdatedAt = date
			testfile.EqualJSON(t, path+"/shared-article.json", art)
			page.ID = 123
			page.CreatedAt = date
			page.UpdatedAt = date
			page.PublicationDate.Time = timex.ToEST(page.PublicationDate.Time)
			testfile.EqualJSON(t, path+"/page.json", page)
		}
	})
}
