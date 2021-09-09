package almanack

import (
	"net/http"

	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/common"
)

type Service struct {
	common.Logger
	Client  *http.Client
	Queries *db.Queries
	common.ContentStore
	ImageStore  aws.BlobStore
	FileStore   aws.BlobStore
	SlackClient slack.Client
	Indexer     index.Indexer
	common.NewletterService
	gsvc *google.Service
}
