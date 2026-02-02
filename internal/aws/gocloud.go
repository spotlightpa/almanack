package aws

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
	"gocloud.dev/blob/s3blob"
)

func register(scheme, region, id, secret string) error {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(id, secret, "")),
	)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	o := &customURLOpener{client: client}
	blob.DefaultURLMux().RegisterBucket(scheme, o)
	return nil
}

// customURLOpener implements blob.BucketURLOpener with static credentials.
type customURLOpener struct {
	client *s3.Client
}

func (o *customURLOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*blob.Bucket, error) {
	return s3blob.OpenBucket(ctx, o.client, u.Host, nil)
}
