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

-- name: UpdateSharedArticleData :one
UPDATE
  shared_article
SET
  raw_data = @raw_data::jsonb
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
