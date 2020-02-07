package aws

import (
	"flag"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/carlmjohnson/crockford"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

func FlagVar(fl *flag.FlagSet) func(l almanack.Logger) almanack.ImageStore {
	// Set the AWS Region that the service clients should use
	region := fl.String("aws-s3-region", "us-east-2", "AWS `region` to use for S3")
	bucket := fl.String("aws-s3-bucket", "", "AWS `bucket` to use for S3")

	return func(l almanack.Logger) almanack.ImageStore {
		cfg, err := external.LoadDefaultAWSConfig()
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

func (is ImageStore) GetSignedUpload() (signedURL, filename string, err error) {
	filename = makeFilename()
	is.l.Printf("creating presigned URL for %q", filename)
	input := &s3.PutObjectInput{
		Bucket: &is.bucket,
		Key:    &filename,
	}
	req := is.svc.PutObjectRequest(input)
	signedURL, err = req.Presign(15 * time.Minute)

	return
}

func makeFilename() string {
	var sb strings.Builder
	sb.Grow(len("2006/01/123456789abcdefg.jpeg"))
	sb.WriteString(time.Now().Format("2006/01/"))
	sb.Write(crockford.Time(crockford.Lower, time.Now()))
	sb.Write(crockford.AppendRandom(crockford.Lower, nil))
	sb.WriteString(".jpeg")
	return sb.String()
}

type MockImageStore struct {
	l almanack.Logger
}

func (mis MockImageStore) GetSignedUpload() (signedURL, filename string, err error) {
	mis.l.Printf("returning mock signed URL")
	return "https://invalid", makeFilename(), nil
}
