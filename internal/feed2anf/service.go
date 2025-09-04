package feed2anf

import "flag"

type Service struct {
	NewsFeedURL string
}

func AddFlags(fl *flag.FlagSet) (svc *Service) {
	svc = new(Service)
	fl.StringVar(&svc.NewsFeedURL, "news-feed-url", "https://www.spotlightpa.org/feeds/full.json", "`URL` for published news feed")
	return svc
}
