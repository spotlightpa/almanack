package db

import (
	"time"

	"github.com/gorilla/feeds"
)

func NewsletterToFeed(title string, archive []Newsletter) feeds.JSONFeed {
	feed := feeds.JSONFeed{
		Title: title,
		Items: make([]*feeds.JSONItem, len(archive)),
	}
	for i, nl := range archive {
		var date time.Time = nl.PublishedAt
		feed.Items[i] = &feeds.JSONItem{
			Id:            nl.ArchiveURL,
			Url:           nl.ArchiveURL,
			Title:         nl.Subject,
			PublishedDate: &date,
		}
	}
	return feed
}
