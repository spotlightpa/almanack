-- name: UpdateSharedArticle :one
UPDATE
  shared_article
SET
  embargo_until = @embargo_until,
  status = @status,
  note = @note
WHERE
  id = @id
RETURNING
  *;

-- name: UpdateSharedArticlePage :one
UPDATE
  shared_article
SET
  page_id = @page_id
WHERE
  id = @shared_article_id
RETURNING
  *;

-- name: GetSharedArticleByID :one
SELECT
  *
FROM
  shared_article
WHERE
  id = $1;

-- name: GetSharedArticleBySource :one
SELECT
  *
FROM
  shared_article
WHERE
  source_type = $1
  AND source_id = $2;

-- name: ListSharedArticles :many
SELECT
  *
FROM
  shared_article
ORDER BY
  updated_at DESC
LIMIT $1 OFFSET $2;

-- name: ListSharedArticlesWhereActive :many
SELECT
  *
FROM
  shared_article
WHERE
  status <> 'U'
ORDER BY
  CASE status
  WHEN 'P' THEN
    '0'
  WHEN 'S' THEN
    '1'
  END ASC,
  updated_at DESC
LIMIT $1 OFFSET $2;

-- name: UpsertSharedArticleFromArc :one
INSERT INTO shared_article (status, source_type, source_id, raw_data)
SELECT
  'U',
  'arc',
  arc.arc_id,
  arc.raw_data
FROM
  arc
WHERE
  arc_id = $1
ON CONFLICT (source_type,
  source_id)
  DO UPDATE SET
    raw_data = excluded.raw_data
  RETURNING
    *;
