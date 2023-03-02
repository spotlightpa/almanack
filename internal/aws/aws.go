package aws

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/carlmjohnson/errorx"
	"gocloud.dev/blob"

	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func AddFlags(fl *flag.FlagSet) func() (imageStore, fileStore BlobStore) {
	accessKeyID := fl.String("aws-access-key", "", "AWS access `key` ID")
	secretAccessKey := fl.String("aws-secret-key", "", "AWS secret access `key`")
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	ibucket := fl.String("image-bucket-url", "mem://", "bucket `URL` for images")
	fbucket := fl.String("file-bucket-url", "mem://", "bucket `URL` for files")

	return func() (imageStore, fileStore BlobStore) {
		err := register("s3-cli", *region, *accessKeyID, *secretAccessKey)
		if err != nil {
			almlog.Logger.Error("aws.register", err)
		}
		imageStore = BlobStore{*ibucket}
		if *ibucket == "mem://" {
			almlog.Logger.Warn("mocking AWS image bucket")
		}
		fileStore = BlobStore{*fbucket}
		if *fbucket == "mem://" {
			almlog.Logger.Warn("mocking AWS file bucket")
		}
		return
	}
}

type BlobStore struct {
	bucket string
}

func (bs BlobStore) SignPutURL(ctx context.Context, srcPath string, h http.Header) (signedURL string, err error) {
	l := almlog.FromContext(ctx)
	l.InfoCtx(ctx, "aws.SignPutURL", "url", srcPath)
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

func (bs BlobStore) SignGetURL(ctx context.Context, srcPath string) (signedURL string, err error) {
	l := almlog.FromContext(ctx)
	l.InfoCtx(ctx, "aws.SignGetURL", "url", srcPath)
	b, err := blob.OpenBucket(ctx, bs.bucket)
	if err != nil {
		return "", err
	}
	defer b.Close()
	return b.SignedURL(ctx, srcPath, &blob.SignedURLOptions{
		Expiry: 24 * time.Hour,
		Method: http.MethodGet,
		BeforeSign: func(as func(any) bool) error {
			var opts *s3.GetObjectInput
			if as(&opts) {
				filename := path.Base(srcPath)
				download := httpx.AttachmentName(filename)
				opts.ResponseContentDisposition = &download
			}
			return nil
		},
	})
}

func (bs BlobStore) BuildURL(srcPath string) string {
	u := must.Get(url.Parse(bs.bucket))

	// Just assuming bucket name is valid DNS…
	return fmt.Sprintf("https://%s/%s", u.Hostname(), srcPath)
}

func (bs BlobStore) WriteFile(ctx context.Context, path string, h http.Header, data []byte) (err error) {
	l := almlog.FromContext(ctx)
	b, err := blob.OpenBucket(ctx, bs.bucket)
	if err != nil {
		return err
	}
	defer errorx.Defer(&err, b.Close)

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
			l.InfoCtx(ctx, "aws.WriteFile: skipping; already uploaded",
				"bucket", bs.bucket, "path", path)
			return nil
		}
	}

	l.InfoCtx(ctx, "aws.WriteFile: writing", "bucket", bs.bucket, "path", path)
	return b.WriteAll(ctx, path, data, &blob.WriterOptions{
		CacheControl:       h.Get("Cache-Control"),
		ContentType:        h.Get("Content-Type"),
		ContentDisposition: h.Get("Content-Disposition"),
		ContentMD5:         checksum,
	})
}
