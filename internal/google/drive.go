package google

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/earthboundkid/bytemap/v2"
)

func (gsvc *Service) DriveClient(ctx context.Context) (cl *http.Client, err error) {
	return gsvc.client(ctx,
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.appdata",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/drive.metadata",
		"https://www.googleapis.com/auth/drive.metadata.readonly",
		"https://www.googleapis.com/auth/drive.photos.readonly",
		"https://www.googleapis.com/auth/drive.readonly",
		"https://www.googleapis.com/auth/drive.scripts",
	)
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
	var v resperr.Validator
	v.AddIf("drive_id", len(s) == 0, "ID must be set")
	id, found := strings.CutPrefix(s, "https://drive.google.com/file/d/")
	v.AddIfUnset("drive_id", !found, "Unrecognized ID prefix: %s", s)
	if found {
		id, _, _ = strings.Cut(id, "/")
	}
	v.AddIfUnset("drive_id", !validIDChars.Contains(id),
		"Illegal characters in file ID: %s", s)
	return id, v.Err()
}

func (gsvc *Service) DownloadURLForDriveID(fileID string) (string, error) {
	id, err := NormalizeFileID(fileID)
	if err != nil {
		return "", err
	}
	u, err := requests.
		URL("https://www.googleapis.com").
		Pathf("/drive/v3/files/%s", id).
		Param("alt", "media").
		URL()
	return u.String(), err
}

func (gsvc *Service) DownloadFile(ctx context.Context, cl *http.Client, fileID string) ([]byte, error) {
	u, err := gsvc.DownloadURLForDriveID(fileID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = requests.
		URL(u).
		Client(cl).
		ToBytesBuffer(&buf).
		Fetch(ctx)
	return buf.Bytes(), err
}
