package db_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/testfile"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
	docs "google.golang.org/api/docs/v1"
)

func TestProcessGDocsDoc(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := context.Background()
	testfile.Run(t, "testdata/gdoc*", func(t *testing.T, path string) {
		t.Parallel()
		svc := almanack.Services{
			Queries:    q,
			Tx:         db.NewTxable(p),
			ImageStore: aws.NewBlobStore("mem://"),
			FileStore:  aws.NewBlobStore("mem://"),
			Gsvc:       new(google.Service),
			Client: &http.Client{
				Transport: requests.Replay(path),
			},
		}
		if os.Getenv("RECORD") != "" {
			svc.Client.Transport = requests.Caching(nil, path)
			cl, _ := svc.Gsvc.DriveClient(context.Background())
			cl.Transport = requests.Caching(cl.Transport, path)
			svc.Gsvc.SetMockClient(cl)
		} else {
			svc.Gsvc.SetMockClient(svc.Client)
		}

		var doc docs.Document
		testfile.ReadJSON(t, path+"/doc.json", &doc)
		// Run twice to test the already uploaded path
		for i := 0; i < 2; i++ {
			dbDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
				ExternalID: fmt.Sprintf("abc123_%s", stringx.Slugify(path)),
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
