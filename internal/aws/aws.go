package aws

import (
	"flag"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/spotlightpa/almanack/pkg/common"
)

func FlagVar(fl *flag.FlagSet) func(l common.Logger) (imageStore, fileStore common.FileStore) {
	accessKeyID := fl.String("aws-access-key", "", "AWS access `key` ID")
	secretAccessKey := fl.String("aws-secret-key", "", "AWS secret access `key`")
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	ibucket := fl.String("aws-s3-bucket", "", "AWS `bucket` to use for S3 images")
	fbucket := fl.String("aws-s3-file-bucket", "", "AWS `bucket` to use for S3 files")

	return func(l common.Logger) (imageStore, fileStore common.FileStore) {
		cfg, err := external.LoadDefaultAWSConfig(
			external.WithCredentialsValue(aws.Credentials{
				AccessKeyID:     *accessKeyID,
				SecretAccessKey: *secretAccessKey,
			}),
		)

		imageStore, fileStore = MockStore{l}, MockStore{l}
		if err != nil {
			l.Printf("using mock AWS: %v", err)
			return
		}
		cfg.Region = *region
		if *ibucket != "" {
			imageStore = S3Store{s3.New(cfg), *ibucket, l}
		} else {
			l.Printf("using mock AWS image bucket")
		}
		if *fbucket != "" {
			fileStore = S3Store{s3.New(cfg), *fbucket, l}
		} else {
			l.Printf("using mock AWS file bucket")
		}
		return
	}
}

type S3Store struct {
	svc    *s3.Client
	bucket string
	l      common.Logger
}

func (ss S3Store) GetSignedURL(srcPath string) (signedURL string, err error) {
	ss.l.Printf("creating presigned URL for %q", srcPath)
	input := &s3.PutObjectInput{
		Bucket: &ss.bucket,
		Key:    &srcPath,
	}
	req := ss.svc.PutObjectRequest(input)
	signedURL, err = req.Presign(15 * time.Minute)

	return
}

type MockStore struct {
	l common.Logger
}

func (ms MockStore) GetSignedURL(srcPath string) (signedURL string, err error) {
	ms.l.Printf("returning mock signed URL")
	return "https://invalid", nil
}
