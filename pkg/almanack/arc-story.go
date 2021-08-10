package almanack

import (
	"errors"

	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
)

type ArcStory struct {
	arc.FeedItem
	Note   string `json:"almanack-note,omitempty"`
	Status Status `json:"almanack-status,omitempty"`
}

func ArcStoryFromDB(dart *db.Article) (story *ArcStory, err error) {
	var newStory ArcStory
	if err = newStory.fromDB(dart); err != nil {
		return
	}
	story = &newStory
	return
}

func (story *ArcStory) fromDB(dart *db.Article) error {
	story.FeedItem = dart.ArcData
	story.Note = dart.Note
	var ok bool
	if story.Status, ok = dbStatusToStatus[dart.Status]; !ok {
		return errors.New("bad status flag in database")
	}
	return nil
}

func storiesFromDB(dbArts []db.Article) ([]ArcStory, error) {
	stories := make([]ArcStory, len(dbArts))
	for i := range stories {
		if err := stories[i].fromDB(&dbArts[i]); err != nil {
			return nil, err
		}
	}
	return stories, nil
}

type Status int8

const (
	StatusUnset     Status = 0
	StatusPlanned   Status = 1
	StatusAvailable Status = 2
)

var dbStatusToStatus = map[string]Status{
	"U": StatusUnset,
	"P": StatusPlanned,
	"A": StatusAvailable,
}

var statusToDBstring = map[Status]string{
	StatusUnset:     "U",
	StatusPlanned:   "P",
	StatusAvailable: "A",
}

func (s Status) dbstring() string {
	return statusToDBstring[s]
}
