package youtube

import (
	"context"
	"encoding/json"
	"time"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/db"
)

func UpdateCache(ctx context.Context, q db.Queries, entries []Entry) (err error) {
	defer errorx.Trace(&err)

	items := convertForDatabase(entries)
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	_, err = q.UpsertYouTubeFeedArchives(ctx, data)
	return
}

type bulkItem struct {
	ExternalID          string    `json:"external_id"`
	Title               string    `json:"title"`
	URL                 string    `json:"url"`
	ThumbnailUrl        string    `json:"thumbnail_url"`
	ExternalPublishedAt time.Time `json:"external_published_at"`
	ExternalUpdatedAt   time.Time `json:"external_updated_at"`
}

func convertForDatabase(entries []Entry) []bulkItem {
	items := make([]bulkItem, len(entries))
	for i, entry := range entries {
		items[i] = bulkItem{
			ExternalID:          entry.ID,
			Title:               entry.Title,
			URL:                 entry.Link.Href,
			ThumbnailUrl:        entry.MediaGroup.Thumbnail.URL,
			ExternalPublishedAt: entry.Published,
			ExternalUpdatedAt:   entry.Updated,
		}
	}
	return items
}
