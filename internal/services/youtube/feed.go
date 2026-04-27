package youtube

import (
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type FeedItem struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Thumbnail string    `json:"thumbnail"`
	PubDate   time.Time `json:"published"`
}

func ToFeed(data []db.Youtube) []FeedItem {
	videos := make([]FeedItem, len(data))
	for i, item := range data {
		videos[i] = FeedItem{
			Title:     item.Title,
			URL:       item.URL,
			Thumbnail: item.ThumbnailUrl,
			PubDate:   item.ExternalPublishedAt,
		}
	}
	return videos
}
