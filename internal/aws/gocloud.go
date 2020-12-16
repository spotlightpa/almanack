package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

func register(scheme, region, id, secret string) error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
	if err != nil {
		return err
	}
	o := s3blob.URLOpener{
		ConfigProvider: sess,
	}
	blob.DefaultURLMux().RegisterBucket(scheme, &o)
	return nil
}
