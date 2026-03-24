package almanack

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlmjohnson/flowmatic"
	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/stringx"
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
	return flowmatic.Do(
		func() error {
			data, err := svc.Queries.ListYouTubeWhereShort(ctx, db.ListYouTubeWhereShortParams{
				Limit:  20,
				Offset: 0,
			})
			if err != nil {
				return err
			}
			return UploadJSON(
				ctx,
				svc.FileStore,
				"feeds/youtube-shorts.json",
				"public, max-age=300",
				struct {
					Videos []youtube.FeedItem `json:"videos"`
				}{
					youtube.ToFeed(data),
				},
			)
		},
		func() error {
			data, err := svc.Queries.ListYouTubeWhereRegular(ctx, db.ListYouTubeWhereRegularParams{
				Limit:  20,
				Offset: 0,
			})
			if err != nil {
				return err
			}
			return UploadJSON(
				ctx,
				svc.FileStore,
				"feeds/youtube-regular.json",
				"public, max-age=300",
				struct {
					Videos []youtube.FeedItem `json:"videos"`
				}{
					youtube.ToFeed(data),
				},
			)
		},
	)
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
	isShort := strings.Contains(video.URL, "/shorts/")
	imageDesc := fmt.Sprint("Video:", video.Title)
	if isShort {
		imageDesc = fmt.Sprint("Short:", video.Title)
	}
	imagePath, err := svc.ReplaceAndUploadImageURL(ctx, video.ThumbnailUrl, imageDesc, "")
	if err != nil {
		return err
	}
	return svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) error {
		defer errorx.Trace(&err)
		path := fmt.Sprintf("content/videos/%s.md", stringx.SlugifyVideoID(video.ExternalID))
		page, err := q.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path,
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
			"internal-id":       stringx.Truncate(imageDesc, 16),
			"published":         video.ExternalPublishedAt,
			"byline":            "",
			"title":             video.Title,
			"blurb":             "",
			"kicker":            "Video",
			"video_url":         video.URL,
			"video_type":        videoType,
			"image":             imagePath,
			"image-description": imageDesc,
		}
		q.UpdatePage(ctx, db.UpdatePageParams{
			ID:               page.ID,
			SetFrontmatter:   true,
			Frontmatter:      fm,
			SetBody:          false,
			SetScheduleFor:   false,
			SetLastPublished: false,
		})
		return err
	})
}
