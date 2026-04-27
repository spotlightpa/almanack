package almanack

import (
	"context"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"github.com/spotlightpa/almanack/internal/youtube"
)

func (svc Services) UpdateYouTubeFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	entries, err := svc.YT.FetchFeed(ctx, svc.Client)
	if err != nil {
		return err
	}
	if err = youtube.UpdateCache(ctx, *svc.Queries, entries); err != nil {
		return err
	}
	if err = svc.Queries.ResetYouTubeMaxID(ctx); err != nil {
		return err
	}
	if err = svc.CreateYouTubePages(ctx); err != nil {
		return err
	}
	return nil
}

func (svc Services) CreateYouTubePages(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	videos, err := svc.Queries.ListYouTubeWhereNoPage(ctx)
	if err != nil {
		return
	}
	for i := range videos {
		video := &videos[i]
		if err = svc.CreateYouTubePage(ctx, video); err != nil {
			return err
		}
	}
	return
}
func (svc Services) CreateYouTubePage(ctx context.Context, video *db.Youtube) (err error) {
	isShort := video.IsShort()
	imageDesc := fmt.Sprintf("Video: %s", video.Title)
	if isShort {
		imageDesc = fmt.Sprintf("Short: %s", video.Title)
	}
	imagePath, err := svc.ReplaceAndUploadImageURL(ctx, video.ThumbnailUrl, imageDesc, "")
	if err != nil {
		return err
	}
	return svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) error {
		defer errorx.Trace(&err)
		page, err := q.CreatePage(ctx, db.CreatePageParams{
			FilePath:   video.FilePath(),
			SourceType: "youtube",
			SourceID:   video.ExternalID,
		})
		if err != nil {
			return err
		}
		pageID := pgtype.Int8{Int64: page.ID, Valid: true}
		_, err = q.UpdateYouTubePage(ctx, db.UpdateYouTubePageParams{
			ID:     video.ID,
			PageID: pageID,
		})
		if err != nil {
			return err
		}
		videoType := "youtube-regular"
		if isShort {
			videoType = "youtube-short"
		}
		fm := db.Map{
			"internal-id":       stringx.Truncate(imageDesc, 20),
			"published":         video.ExternalPublishedAt,
			"byline":            "",
			"title":             video.Title,
			"blurb":             "",
			"kicker":            "",
			"youtube-id":        video.YouTubeID(),
			"link":              video.URL,
			"video-url":         video.URL,
			"video-type":        videoType,
			"image":             imagePath,
			"image-description": imageDesc,
			"draft":             false,
		}
		page, err = q.UpdatePage(ctx, db.UpdatePageParams{
			ID:               page.ID,
			SetFrontmatter:   true,
			Frontmatter:      fm,
			SetBody:          false,
			SetScheduleFor:   false,
			SetLastPublished: true, // Rolls back on error
		})
		if err != nil {
			return err
		}

		data, err := page.ToJSON()
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("Content: publishing %q", stringx.Truncate(imageDesc, 25))
		if err = svc.ContentStore.UpdateFile(ctx, msg, page.FilePath, data); err != nil {
			return err
		}

		return nil
	})
}
