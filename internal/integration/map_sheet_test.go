package integration_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/jackc/pgx/v5"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/google"
	"github.com/spotlightpa/almanack/internal/services/index"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

const (
	searchableSheetID = "searchable0000000000000000000000000000000000"
	scrollySheetID    = "scrolly000000000000000000000000000000000000x"
)

func TestMapSheetPage(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	ctx := t.Context()

	tmp := t.ArtifactDir()
	svc := almsvc.Services{
		DB:           dbhandle,
		Queries:      dbhandle.Queries(),
		Auth:         netlifyid.MockAuthService{},
		Gsvc:         new(google.Service),
		ContentStore: github.NewMockClient(tmp),
		Indexer:      index.MockIndexer{},
	}
	svc.Gsvc.SetMockClient(&http.Client{Transport: reqtest.Replay("testdata/map sheet")})
	_ = be.OK(svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:   "2026/11/share.jpeg",
		Type:   "jpeg",
		Credit: "Photographer / Spotlight PA",
	}))
	_ = be.OK(svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:   "2026/11/stop.jpeg",
		Type:   "jpeg",
		Width:  1600,
		Height: 900,
	}))
	rb := newTestServer(t, svc)

	var info struct {
		SheetID       string `json:"sheet_id"`
		FeaturedDocID string `json:"featured_doc_id"`
	}
	be.NilError(rb.Clone().
		Path("/api/map-sheet").
		Param("id", "https://docs.google.com/spreadsheets/d/"+searchableSheetID+"/edit#gid=0").
		ToJSON(&info).
		Fetch(ctx))
	be.Equal(info.SheetID, searchableSheetID).
		Equal(info.FeaturedDocID, "")

	var pageID int64
	be.NilError(rb.Clone().
		Path("/api/map-sheet").
		BodyJSON(map[string]string{"id": searchableSheetID}).
		ToJSON(&pageID).
		Fetch(ctx))

	page := be.OK(svc.Queries.GetPageByID(ctx, pageID))
	be.Equal(page.FilePath, "content/news/2026-11-13-SPLTESTMAP.md").
		Equal(page.SourceType, almsvc.MapSheetSourceType).
		Equal(page.SourceID, searchableSheetID).
		False(page.LastPublished.Valid).
		Equal(page.Frontmatter["layout"], "searchable-map").
		Equal(page.Frontmatter["image-credit"], "Photographer / Spotlight PA").
		True(strings.HasPrefix(page.Body, `{{<featured/map-header eyebrow="Interactive"`)).
		True(strings.Contains(page.Body, "Intro text\n{{</featured/map-header>}}")).
		True(strings.Contains(page.Body, `{{<featured/credit eyebrow="Reporting"`))

	var againID int64
	be.NilError(rb.Clone().
		Path("/api/map-sheet").
		BodyJSON(map[string]string{"id": searchableSheetID}).
		ToJSON(&againID).
		Fetch(ctx))
	be.Equal(againID, pageID)

	_ = be.OK(svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		ID:             pageID,
		SetFrontmatter: true,
		Frontmatter:    db.Map{"title": "Edited in Almanack", "published": page.Frontmatter["published"]},
		SetBody:        true,
		Body:           "stale",
		URLPath:        "",
	}))

	var refreshed db.Page
	be.NilError(rb.Clone().
		Path("/api/page-refresh").
		BodyJSON(map[string]any{"id": strconv.FormatInt(pageID, 10), "refresh_metadata": false}).
		ToJSON(&refreshed).
		Fetch(ctx))
	be.Equal(refreshed.Frontmatter["title"], "Edited in Almanack").
		True(strings.HasPrefix(refreshed.Body, "{{<featured/map-header"))

	be.NilError(rb.Clone().
		Path("/api/page-refresh").
		BodyJSON(map[string]any{"id": strconv.FormatInt(pageID, 10), "refresh_metadata": true}).
		ToJSON(&refreshed).
		Fetch(ctx))
	be.Equal(refreshed.Frontmatter["title"], "Find Your District").
		Equal(refreshed.Frontmatter["layout"], "searchable-map")

	be.NilError(svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) error {
		err, warning := svc.PublishPage(ctx, txq, &refreshed)
		be.NilError(warning)
		return err
	}))
	published := string(be.OK(os.ReadFile(filepath.Join(tmp, page.FilePath))))
	be.True(strings.HasPrefix(published, "+++\n")).
		True(strings.Contains(published, `layout = "searchable-map"`)).
		True(strings.Contains(published, "+++\n\n{{<featured/map-header"))
	page = be.OK(svc.Queries.GetPageByID(ctx, pageID))
	be.True(page.LastPublished.Valid).
		Equal(page.URLPath.String, "/news/2026/11/test-district-map/")

	be.NilError(rb.Clone().
		Path("/api/map-sheet").
		BodyJSON(map[string]string{"id": scrollySheetID}).
		ToJSON(&pageID).
		Fetch(ctx))
	page = be.OK(svc.Queries.GetPageByID(ctx, pageID))
	be.Equal(page.FilePath, "content/news/2026-11-14-SPLTESTSCROLLY.md").
		Equal(page.Frontmatter["layout"], "scrolly-map").
		True(strings.Contains(page.Body, `image-ratio="16:9"`)).
		True(strings.Contains(page.Body, `label="Coal Region"`)).
		Equal(strings.Count(page.Body, "{{<featured/scrolly-map-stop"), 2)

	err := rb.Clone().
		Path("/api/map-sheet").
		BodyJSON(map[string]string{"id": "not-a-sheet"}).
		Fetch(ctx)
	be.True(err != nil)
}
