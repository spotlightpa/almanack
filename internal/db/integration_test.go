package db_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
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

func TestEmbed_UnmarshalJSON(t *testing.T) {
	{
		e1 := db.Embed{
			N:    1,
			Type: db.ImageEmbedTag,
			Value: db.EmbedImage{
				Path:        "path",
				Credit:      "credit",
				Caption:     "caption",
				Description: "desc",
			},
		}
		b, err := json.Marshal(e1)
		be.NilErr(t, err)
		var e2 db.Embed
		be.NilErr(t, json.Unmarshal(b, &e2))
		be.Equal(t, e1, e2)
	}
	{
		e1 := db.Embed{
			N:     2,
			Type:  db.RawEmbedTag,
			Value: "Mork from Ork",
		}
		b, err := json.Marshal(e1)
		be.NilErr(t, err)
		var e2 db.Embed
		be.NilErr(t, json.Unmarshal(b, &e2))
		be.Equal(t, e1, e2)
	}
	{
		e1 := db.Embed{
			Type: "bad",
		}
		b, err := json.Marshal(e1)
		be.NilErr(t, err)
		var e2 db.Embed
		be.Nonzero(t, json.Unmarshal(b, &e2))
	}
}

func TestPublishAppleNews(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	cl := &http.Client{
		Transport: reqtest.Caching(almlog.HTTPTransport, "testdata/anf"),
	}
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("used default client"))
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		NewsFeed: &jsonfeed.NewsFeed{
			URL: "https://www.spotlightpa.org/feeds/full.json",
		},
	}
	be.NilErr(t, svc.UpdateAppleNewsArchive(ctx))
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 15, newItems)

	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))
	// Publishing should mark everyone as uploaded
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.Zero(t, newItems)

	// Updating archive should not mark previously uploaded items as null
	be.NilErr(t, svc.UpdateAppleNewsArchive(ctx))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 0, newItems)
}
