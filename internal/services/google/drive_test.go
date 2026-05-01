package google

import (
	"cmp"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
)

func TestListDriveFiles(t *testing.T) {
	svc := Service{}
	svc.driveID = cmp.Or(os.Getenv("ALMANACK_GOOGLE_TEST_DRIVE"), "1")
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.DriveClient(ctx)
		if err != nil {
			t.Fatal(err)
		}
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	files, err := svc.Files(ctx, &cl)
	be.NilErr(t, err)
	be.Nonzero(t, files)
}

func TestDownloadFile(t *testing.T) {
	var gsvc Service
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	b, err := gsvc.DownloadFile(ctx, &cl, "1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL")
	be.NilErr(t, err)
	be.Equal(t, "image/jpeg", http.DetectContentType(b))

	b, err = gsvc.DownloadFile(ctx, &cl, "https://drive.google.com/file/d/1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL/view?usp=share_link")
	be.NilErr(t, err)
	be.Equal(t, "image/jpeg", http.DetectContentType(b))

	b, err = gsvc.DownloadFile(ctx, &cl, "https://drive.google.com/file/d/1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL;;/view?usp=share_link")
	be.Nonzero(t, err)
	be.Zero(t, b)
}
