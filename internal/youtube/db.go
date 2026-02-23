package youtube

import "time"

type BulkItem struct {
	ExternalID          string    `json:"external_id"`
	Title               string    `json:"title"`
	URL                 string    `json:"url"`
	ThumbnailUrl        string    `json:"thumbnail_url"`
	ExternalPublishedAt time.Time `json:"external_published_at"`
	ExternalUpdatedAt   time.Time `json:"external_updated_at"`
}
