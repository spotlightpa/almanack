package index

import (
	"context"
	"flag"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func AddFlags(fl *flag.FlagSet) func() Indexer {
	appID := fl.String("indexer-app-id", "", "`app id` for Algolia")
	apiKey := fl.String("indexer-api-key", "", "`api key` for Algolia")
	indexName := fl.String("indexer-index-name", "", "`index` name for Algolia")
	return func() Indexer {
		if *apiKey == "" {
			almlog.Logger.Warn("mocking indexer")
			return MockIndexer{}
		}

		client := search.NewClient(*appID, *apiKey)
		index := client.InitIndex(*indexName)
		return index
	}
}

type Indexer interface {
	SaveObject(object any, opts ...any) (res search.SaveObjectRes, err error)
}

type MockIndexer struct {
}

func (mi MockIndexer) SaveObject(object any, opts ...any) (res search.SaveObjectRes, err error) {
	l := almlog.Logger
	var ctx context.Context
	var ok bool
	for _, opt := range opts {
		if ctx, ok = opt.(context.Context); ok {
			l = almlog.FromContext(ctx)
		}
	}
	l.InfoContext(ctx, "index.Mock.SaveObject")
	return
}
