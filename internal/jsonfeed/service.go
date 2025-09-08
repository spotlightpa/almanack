package jsonfeed

import "flag"

type NewsFeed struct {
	URL string
}

func AddFlags(fl *flag.FlagSet) (svc *NewsFeed) {
	svc = new(NewsFeed)
	fl.StringVar(&svc.URL, "news-feed-url", "https://www.spotlightpa.org/feeds/full.json", "`URL` for published news feed")
	return svc
}
