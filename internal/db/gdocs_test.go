package db_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/testfile"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
	docs "google.golang.org/api/docs/v1"
)

func TestProcessGDocsDoc(t *testing.T) {
	almlog.UseDevLogger()
	p := createTestDB(t)
	q := db.New(p)
	svc := almanack.Services{
		Queries:    q,
		Tx:         db.NewTxable(p),
		ImageStore: aws.NewBlobStore("mem://"),
		FileStore:  aws.NewBlobStore("mem://"),
	}
	ctx := context.Background()
	testfile.GlobRun(t, "testdata/gdoc*", func(path string, t *testing.T) {
		svc.Client = &http.Client{
			Transport: requests.Replay(path),
		}
		var doc docs.Document
		testfile.ReadJSON(t, path+"/doc.json", &doc)
		// Run twice to test the already uploaded path
		for i := 0; i < 2; i++ {
			dbDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
				GDocsID:  fmt.Sprintf("abc123_%s", stringx.Slugify(path)),
				Document: doc,
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
			gotWarnings := strings.Join(dbDoc.Warnings, "\n")
			testfile.Equal(rt, path+"/warnings.txt", gotWarnings)
		}
	})
}
