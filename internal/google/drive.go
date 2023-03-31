package google

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/bytemap"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (gsvc *Service) DriveClient(ctx context.Context) (cl *http.Client, err error) {
	scopes := []string{
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.appdata",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/drive.metadata",
		"https://www.googleapis.com/auth/drive.metadata.readonly",
		"https://www.googleapis.com/auth/drive.photos.readonly",
		"https://www.googleapis.com/auth/drive.readonly",
		"https://www.googleapis.com/auth/drive.scripts",
	}
	if len(gsvc.cert) == 0 {
		l := almlog.FromContext(ctx)
		l.WarnCtx(ctx, "using default Google credentials")
		cl, err = google.DefaultClient(ctx, scopes...)
		return
	}
	creds, err := google.CredentialsFromJSON(ctx, gsvc.cert, scopes...)
	if err != nil {
		return
	}
	cl = oauth2.NewClient(ctx, creds.TokenSource)
	return
}

func (gsvc *Service) Files(ctx context.Context, cl *http.Client) (files []*Files, err error) {
	var fileList FileList
	if err = requests.
		URL("https://www.googleapis.com/drive/v3/files").
		Param("corpora", "drive").
		Param("driveId", gsvc.driveID).
		Param("includeItemsFromAllDrives", "true").
		Param("supportsAllDrives", "true").
		Param("q", "mimeType='application/vnd.google-apps.document'").
		Param("fields", "files(id,description,lastModifyingUser,name)").
		Param("orderBy", "recency").
		Param("pageSize", "50").
		Client(cl).
		ToJSON(&fileList).
		Fetch(ctx); err != nil {
		return nil, err
	}
	files = fileList.Files
	return
}

type FileList struct {
	Files []*Files `json:"files"`
}

type Files struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	LastModifyingUser LastModifyingUser `json:"lastModifyingUser"`
}

type LastModifyingUser struct {
	Kind         string `json:"kind"`
	DisplayName  string `json:"displayName"`
	PhotoLink    string `json:"photoLink"`
	Me           bool   `json:"me"`
	PermissionID string `json:"permissionId"`
	EmailAddress string `json:"emailAddress"`
}

var validIDChars = bytemap.Union(
	bytemap.Range('A', 'Z'),
	bytemap.Range('a', 'z'),
	bytemap.Range('0', '9'),
	bytemap.Make("_-"),
)

func NormalizeFileID(s string) (string, error) {
	// E.g. https://drive.google.com/file/d/<ID>/view?usp=share_link
	if validIDChars.Contains(s) {
		return s, nil
	}
	trimmed, ok := strings.CutPrefix(s, "https://drive.google.com/file/d/")
	if !ok {
		return "", fmt.Errorf("bad file ID prefix: %q", s)
	}
	trimmed, _, _ = strings.Cut(trimmed, "/")
	if !validIDChars.Contains(trimmed) {
		return "", fmt.Errorf("file ID contains illegal characters: %q", s)
	}
	return trimmed, nil
}

func (gsvc Service) DownloadFile(ctx context.Context, cl *http.Client, fileID string) ([]byte, error) {
	if !validIDChars.Contains(fileID) {
		return nil, fmt.Errorf("bad fileID: %q", fileID)
	}
	var buf bytes.Buffer
	err := requests.
		URL("https://www.googleapis.com").
		Pathf("/drive/v3/files/%s", fileID).
		Param("alt", "media").
		Client(cl).
		ToBytesBuffer(&buf).
		Fetch(ctx)
	return buf.Bytes(), err
}
