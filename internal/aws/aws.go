package aws

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"gocloud.dev/blob"

	"github.com/spotlightpa/almanack/pkg/common"
)

func FlagVar(fl *flag.FlagSet) func(l common.Logger) (imageStore, fileStore common.FileStore) {
	accessKeyID := fl.String("aws-access-key", "", "AWS access `key` ID")
	secretAccessKey := fl.String("aws-secret-key", "", "AWS secret access `key`")
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	ibucket := fl.String("image-bucket-url", "mem://", "bucket `URL` for images")
	fbucket := fl.String("file-bucket-url", "mem://", "bucket `URL` for files")

	return func(l common.Logger) (imageStore, fileStore common.FileStore) {
		err := register("s3-cli", *region, *accessKeyID, *secretAccessKey)
		if err != nil {
			l.Printf("problem registering gocloud: %v", err)
		}
		imageStore = S3Store{*ibucket, l}
		if *ibucket == "mem://" {
			l.Printf("using mock AWS image bucket")
		}
		fileStore = S3Store{*fbucket, l}
		if *fbucket == "mem://" {
			l.Printf("using mock AWS file bucket")
		}
		return
	}
}

type S3Store struct {
	bucket string
	l      common.Logger
}

func (ss S3Store) GetSignedURL(srcPath string, h http.Header) (signedURL string, err error) {
	ss.l.Printf("creating presigned URL for %q", srcPath)
	ctx := context.TODO()
	b, err := blob.OpenBucket(ctx, ss.bucket)
	if err != nil {
		return "", err
	}
	defer b.Close()
	return b.SignedURL(ctx, srcPath, &blob.SignedURLOptions{
		Expiry:                   15 * time.Minute,
		Method:                   http.MethodPut,
		ContentType:              h.Get("Content-Type"),
		EnforceAbsentContentType: true,
		BeforeSign: func(as func(interface{}) bool) error {
			var opts *s3.PutObjectInput
			if as(&opts) {
				if disposition := h.Get("Content-Disposition"); disposition != "" {
					opts.ContentDisposition = &disposition
				}
				if cc := h.Get("Cache-Control"); cc != "" {
					opts.CacheControl = &cc
				}
			}
			return nil
		},
	})
}

func (ss S3Store) BuildURL(srcPath string) string {
	u, err := url.Parse(ss.bucket)
	if err != nil {
		panic(err)
	}

	// Just assuming bucket name is valid DNS…
	return fmt.Sprintf("https://%s/%s", u.Hostname(), srcPath)
}
