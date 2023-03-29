package db_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
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
		testfile.Unmarshal(t, path+"/doc.json", &doc)
		dbDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
			GDocsID:  "abc123",
			Document: doc,
		})
		be.NilErr(t, err)
		err = svc.ProcessGDocsDoc(ctx, dbDoc)
		be.NilErr(t, err)
		dbDoc, err = svc.Queries.GetGDocsByID(ctx, dbDoc.ID)
		be.NilErr(t, err)

		rawHTML := testfile.Read(be.Relaxed(t), path+"/raw.html")
		be.Debug(t, func() {
			if rawHTML != dbDoc.RawHtml {
				testfile.Write(t, path+"/raw-bad.html", dbDoc.RawHtml)
			}
		})
		be.Equal(t, rawHTML, dbDoc.RawHtml)

		richText := testfile.Read(be.Relaxed(t), path+"/rich.html")
		be.Debug(t, func() {
			if richText != dbDoc.RichText {
				testfile.Write(t, path+"/rich-bad.html", dbDoc.RichText)
			}
		})
		be.Equal(t, richText, dbDoc.RichText)

		md := testfile.Read(be.Relaxed(t), path+"/article.md")
		be.Debug(t, func() {
			testfile.Write(t, path+"/article-bad.md", dbDoc.ArticleMarkdown)
		})
		be.Equal(t, md, dbDoc.ArticleMarkdown)
	})
}
