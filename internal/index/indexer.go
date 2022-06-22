package index

import (
	"flag"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

func AddFlags(fl *flag.FlagSet) func(Logger) Indexer {
	if fl == nil {
		fl = flag.CommandLine
	}
	appID := fl.String("indexer-app-id", "", "`app id` for Algolia")
	apiKey := fl.String("indexer-api-key", "", "`api key` for Algolia")
	indexName := fl.String("indexer-index-name", "", "`index` name for Algolia")
	return func(l Logger) Indexer {
		if *apiKey == "" {
			l.Printf("using mock indexer")
			return MockIndexer{l: l}
		}

		client := search.NewClient(*appID, *apiKey)
		index := client.InitIndex(*indexName)
		return index
	}
}

type Logger interface {
	Printf(string, ...any)
}

type Indexer interface {
	SaveObject(object any, opts ...any) (res search.SaveObjectRes, err error)
}

type MockIndexer struct {
	l Logger
}

func (mi MockIndexer) SaveObject(object any, opts ...any) (res search.SaveObjectRes, err error) {
	mi.l.Printf("mock indexing")
	return
}
