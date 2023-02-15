package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const listAllPages = `-- name: ListAllPages :many
SELECT
  id,
  file_path,
  coalesce(frontmatter ->> 'internal-id', '')::text AS internal_id,
  coalesce(frontmatter ->> 'title', '')::text AS hed,
  ARRAY (
    SELECT
      jsonb_array_elements_text(
        CASE WHEN frontmatter ->> 'authors' IS NOT NULL THEN
          frontmatter -> 'authors'
        ELSE
          '[]'::jsonb
        END))::text[] AS authors,
  published_at::timestamptz AS pub_date
FROM
  page
WHERE
  published_at IS NOT NULL
ORDER BY
  published_at DESC
`

type ListAllPagesRow struct {
	ID         int64                    `json:"id"`
	FilePath   string                   `json:"file_path"`
	InternalID string                   `json:"internal_id"`
	Hed        string                   `json:"hed"`
	Authors    pgtype.FlatArray[string] `json:"authors"`
	PubDate    time.Time                `json:"pub_date"`
}

func (q *Queries) ListAllPages(ctx context.Context) ([]ListAllPagesRow, error) {
	rows, err := q.db.Query(ctx, listAllPages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllPagesRow
	for rows.Next() {
		var i ListAllPagesRow
		if err := rows.Scan(
			&i.ID,
			&i.FilePath,
			&i.InternalID,
			&i.Hed,
			&i.Authors,
			&i.PubDate,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
