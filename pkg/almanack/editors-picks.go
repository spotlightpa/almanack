package almanack

import (
	"context"
	"encoding/json"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/common"
)

const EditorsPicksLoc = "data/editorsPicks.json"

type EditorsPicks struct {
	FeaturedStories  []string `json:"featuredStories"`
	Subfeatures      []string `json:"subfeatures"`
	LimitSubfeatures bool     `json:"limitSubfeatures"`
	SubfeaturesLimit int      `json:"subfeaturesLimit"`
	TopSlots         []string `json:"topSlots"`
	SidebarPicks     []string `json:"sidebarPicks"`
}

func GetEditorsPicks(ctx context.Context, q *db.Queries) (picks *EditorsPicks, err error) {
	raw, err := q.GetSiteData(ctx, EditorsPicksLoc)
	if err != nil {
		return
	}
	var val EditorsPicks
	if err = json.Unmarshal(raw, &val); err != nil {
		return
	}
	picks = &val
	return
}

func SetEditorsPicks(ctx context.Context, q *db.Queries, gh common.ContentStore, picks *EditorsPicks) (err error) {
	raw, err := json.MarshalIndent(picks, "", "  ")
	if err != nil {
		return err
	}
	if err = q.SetSiteData(ctx, db.SetSiteDataParams{
		Key:  EditorsPicksLoc,
		Data: raw,
	}); err != nil {
		return err
	}
	return gh.UpdateFile(ctx, "Setting Editor's Picks", EditorsPicksLoc, raw)
}
