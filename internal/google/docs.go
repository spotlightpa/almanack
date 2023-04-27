package google

import (
	"context"
	"net/http"
)

func (gsvc *Service) GDocsClient(ctx context.Context) (cl *http.Client, err error) {
	return gsvc.client(ctx,
		"https://www.googleapis.com/auth/documents.readonly",
		"https://www.googleapis.com/auth/drive.readonly",
	)
}
