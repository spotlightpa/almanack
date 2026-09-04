package google

import (
	"cmp"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"net/http"
	"os"
	"testing"
)

func TestListDriveFiles(t *testing.T) {
	be := assert.FailNow(t)
	svc := Service{
		driveID: cmp.Or(os.Getenv("ALMANACK_GOOGLE_TEST_DRIVE"), "1")}
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
	files := be.OK(svc.Files(ctx, &cl))
	be.NotZero(files)
}

func TestDownloadFile(t *testing.T) {
	be := assert.FailNow(t)
	var gsvc Service
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	b := be.OK(gsvc.DownloadFile(ctx, &cl, "1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL"))
	be.Equal(http.DetectContentType(b), "image/jpeg")

	b, err := gsvc.DownloadFile(ctx, &cl, "https://drive.google.com/file/d/1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL/view?usp=share_link")
	be.
		Zero(err).
		Equal(http.DetectContentType(b), "image/jpeg")

	b, err = gsvc.DownloadFile(ctx, &cl, "https://drive.google.com/file/d/1ssiQd8AKXHo99qkZZwYbHxfVJHY3RPnL;;/view?usp=share_link")
	be.
		NotZero(err).
		Zero(b)
}
