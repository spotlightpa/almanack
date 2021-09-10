-- name: UpdateArcArticles :exec
WITH arc_table AS (
  SELECT
    jsonb_array_elements(@arc_items::jsonb) AS article_data)
INSERT INTO article (arc_id, arc_data)
SELECT
  article_data ->> '_id',
  article_data
FROM
  arc_table
ON CONFLICT (arc_id)
  DO UPDATE SET
    arc_data = excluded.arc_data
  WHERE
    article.status = 'U';

-- name: UpdateAlmanackArticle :one
UPDATE
  article
SET
  status = @status,
  note = @note,
  arc_data = CASE WHEN @set_arc_data::bool THEN
    @arc_data::jsonb
  ELSE
    arc_data
  END
WHERE
  arc_id = @arc_id::text
RETURNING
  *;

-- name: GetArticleByArcID :one
SELECT
  *
FROM
  article
WHERE
  arc_id = @arc_id::text;

-- name: ListUpcoming :many
SELECT
  *
FROM
  article
ORDER BY
  arc_data ->> 'last_updated_date' DESC;

-- name: ListAllArcArticles :many
SELECT
  *
FROM
  article
WHERE
  arc_id IS NOT NULL
ORDER BY
  arc_data ->> 'last_updated_date' DESC
LIMIT $1 OFFSET $2;

-- name: ListAvailableArticles :many
SELECT
  *
FROM
  article
WHERE
  status <> 'U'
ORDER BY
  CASE status
  WHEN 'P' THEN
    '0'
  WHEN 'A' THEN
    '1'
  END ASC,
  arc_data ->> 'last_updated_date' DESC
LIMIT $1 OFFSET $2;

-- name: UpdateArcArticleSpotlightPAPath :one
UPDATE
  article
SET
  spotlightpa_path = @spotlightpa_path::text
WHERE
  arc_id = @arc_id::text
RETURNING
  *;
