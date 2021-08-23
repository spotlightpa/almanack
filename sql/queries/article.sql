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
  arc_id = @arc_id
RETURNING
  *;

-- name: UpdateSpotlightPAArticle :one
UPDATE
  article AS "new"
SET
  spotlightpa_data = @spotlightpa_data,
  schedule_for = @schedule_for,
  spotlightpa_path = CASE WHEN "new".spotlightpa_path IS NULL THEN
    @spotlightpa_path
  ELSE
    "new".spotlightpa_path
  END
FROM
  article AS "old"
WHERE
  "new".id = "old".id
  AND "old".arc_id = @arc_id
RETURNING
  "old".schedule_for;

-- name: UpdateSpotlightPAArticleLastPublished :one
UPDATE
  article AS "new"
SET
  last_published = CURRENT_TIMESTAMP
FROM
  article AS "old"
WHERE
  "new".id = "old".id
  AND "old".arc_id = @arc_id::text
RETURNING
  "old".last_published;

-- name: PopScheduled :many
UPDATE
  article
SET
  last_published = CURRENT_TIMESTAMP
WHERE
  last_published IS NULL
  AND schedule_for < (CURRENT_TIMESTAMP + '5 minutes'::interval)
RETURNING
  *;

-- name: GetArticleByArcID :one
SELECT
  *
FROM
  article
WHERE
  arc_id = @arc_id::text;

-- name: GetArticleByDBID :one
SELECT
  *
FROM
  article
WHERE
  id = $1;

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
