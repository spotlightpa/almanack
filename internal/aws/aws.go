package aws

import (
	"flag"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

func FlagVar(fl *flag.FlagSet) func(l almanack.Logger) almanack.ImageStore {
	accessKeyID := fl.String("aws-access-key", "", "AWS access `key` ID")
	secretAccessKey := fl.String("aws-secret-key", "", "AWS secret access `key`")
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	bucket := fl.String("aws-s3-bucket", "", "AWS `bucket` to use for S3")

	return func(l almanack.Logger) almanack.ImageStore {
		cfg, err := external.LoadDefaultAWSConfig(
			external.WithCredentialsValue(aws.Credentials{
				AccessKeyID:     *accessKeyID,
				SecretAccessKey: *secretAccessKey,
			}),
		)

		if err != nil || *bucket == "" {
			l.Printf("using mock AWS: %v", err)
			return MockImageStore{l}
		}
		cfg.Region = *region
		return ImageStore{s3.New(cfg), *bucket, l}
	}
}

type ImageStore struct {
	svc    *s3.Client
	bucket string
	l      almanack.Logger
}

func (is ImageStore) GetSignedURL(srcPath string) (signedURL string, err error) {
	is.l.Printf("creating presigned URL for %q", srcPath)
	input := &s3.PutObjectInput{
		Bucket: &is.bucket,
		Key:    &srcPath,
	}
	req := is.svc.PutObjectRequest(input)
	signedURL, err = req.Presign(15 * time.Minute)

	return
}

type MockImageStore struct {
	l almanack.Logger
}

func (mis MockImageStore) GetSignedURL(srcPath string) (signedURL string, err error) {
	mis.l.Printf("returning mock signed URL")
	return "https://invalid", nil
}
