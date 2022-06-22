package aws

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/carlmjohnson/errutil"
	"gocloud.dev/blob"

	"github.com/spotlightpa/almanack/pkg/common"
)

func AddFlags(fl *flag.FlagSet) func(l common.Logger) (imageStore, fileStore BlobStore) {
	accessKeyID := fl.String("aws-access-key", "", "AWS access `key` ID")
	secretAccessKey := fl.String("aws-secret-key", "", "AWS secret access `key`")
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	ibucket := fl.String("image-bucket-url", "mem://", "bucket `URL` for images")
	fbucket := fl.String("file-bucket-url", "mem://", "bucket `URL` for files")

	return func(l common.Logger) (imageStore, fileStore BlobStore) {
		err := register("s3-cli", *region, *accessKeyID, *secretAccessKey)
		if err != nil {
			l.Printf("problem registering gocloud: %v", err)
		}
		imageStore = BlobStore{*ibucket, l}
		if *ibucket == "mem://" {
			l.Printf("using mock AWS image bucket")
		}
		fileStore = BlobStore{*fbucket, l}
		if *fbucket == "mem://" {
			l.Printf("using mock AWS file bucket")
		}
		return
	}
}

type BlobStore struct {
	bucket string
	l      common.Logger
}

func (bs BlobStore) GetSignedURL(ctx context.Context, srcPath string, h http.Header) (signedURL string, err error) {
	bs.l.Printf("creating presigned URL for %q", srcPath)
	b, err := blob.OpenBucket(ctx, bs.bucket)
	if err != nil {
		return "", err
	}
	defer b.Close()
	return b.SignedURL(ctx, srcPath, &blob.SignedURLOptions{
		Expiry:                   15 * time.Minute,
		Method:                   http.MethodPut,
		ContentType:              h.Get("Content-Type"),
		EnforceAbsentContentType: true,
		BeforeSign: func(as func(any) bool) error {
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

func (bs BlobStore) BuildURL(srcPath string) string {
	u, err := url.Parse(bs.bucket)
	if err != nil {
		panic(err)
	}

	// Just assuming bucket name is valid DNS…
	return fmt.Sprintf("https://%s/%s", u.Hostname(), srcPath)
}

func (bs BlobStore) WriteFile(ctx context.Context, path string, h http.Header, data []byte) (err error) {
	b, err := blob.OpenBucket(ctx, bs.bucket)
	if err != nil {
		return err
	}
	defer errutil.Defer(&err, b.Close)

	var checksum []byte

	// If attrs + MD5 match skip
	if attrs, err := b.Attributes(ctx, path); err == nil &&
		attrs.MD5 != nil &&
		attrs.CacheControl == h.Get("Cache-Control") &&
		attrs.ContentType == h.Get("Content-Type") &&
		attrs.ContentDisposition == h.Get("Content-Disposition") {
		// Get checksum
		a := md5.Sum(data)
		checksum = a[:]
		if string(checksum) == string(attrs.MD5) {
			bs.l.Printf("skipping %q %q; already uploaded", bs.bucket, path)
			return nil
		}
	}

	bs.l.Printf("writing to %q %q", bs.bucket, path)
	return b.WriteAll(ctx, path, data, &blob.WriterOptions{
		CacheControl:       h.Get("Cache-Control"),
		ContentType:        h.Get("Content-Type"),
		ContentDisposition: h.Get("Content-Disposition"),
		ContentMD5:         checksum,
	})
}
