-- name: EnsurePage :exec
INSERT INTO page ("file_path")
  VALUES (@file_path)
ON CONFLICT (file_path)
  DO NOTHING;

-- name: UpdatePage :one
UPDATE
  page
SET
  frontmatter = CASE WHEN @set_frontmatter::boolean THEN
    @frontmatter
  ELSE
    frontmatter
  END,
  body = CASE WHEN @set_body::boolean THEN
    @body
  ELSE
    body
  END,
  schedule_for = CASE WHEN @set_schedule_for::boolean THEN
    @schedule_for
  ELSE
    schedule_for
  END,
  url_path = CASE WHEN @url_path::text != '' THEN
    @url_path
  ELSE
    url_path
  END,
  last_published = CASE WHEN @set_last_published::boolean THEN
    CURRENT_TIMESTAMP
  ELSE
    last_published
  END
WHERE
  file_path = @file_path
RETURNING
  *;

-- name: GetPageByPath :one
SELECT
  *
FROM
  "page"
WHERE
  file_path = $1;

-- name: GetPageByID :one
SELECT
  *
FROM
  "page"
WHERE
  id = $1;

-- name: PopScheduledPages :many
UPDATE
  page
SET
  last_published = CURRENT_TIMESTAMP
WHERE
  last_published IS NULL
  AND schedule_for < (CURRENT_TIMESTAMP + '5 minutes'::interval)
RETURNING
  *;

-- Cannot use coalesce, see https://github.com/kyleconroy/sqlc/issues/780.
-- Treating published_at as text because it sorts faster and we don't do
-- date stuff on the backend, just frontend.
-- name: ListPages :many
SELECT
  "id",
  "file_path",
  (
    CASE WHEN frontmatter ->> 'internal-id' IS NOT NULL THEN
      frontmatter ->> 'internal-id'
    ELSE
      ''
    END)::text AS "internal_id",
  (
    CASE WHEN frontmatter ->> 'title' IS NOT NULL THEN
      frontmatter ->> 'title'
    ELSE
      ''
    END)::text AS "title",
  (
    CASE WHEN frontmatter ->> 'description' IS NOT NULL THEN
      frontmatter ->> 'description'
    ELSE
      ''
    END)::text AS "description",
  (
    CASE WHEN frontmatter ->> 'blurb' IS NOT NULL THEN
      frontmatter ->> 'blurb'
    ELSE
      ''
    END)::text AS "blurb",
  (
    CASE WHEN frontmatter ->> 'image' IS NOT NULL THEN
      frontmatter ->> 'image'
    ELSE
      ''
    END)::text AS "image",
  coalesce("url_path", ''),
  "last_published",
  "created_at",
  "updated_at",
  "schedule_for",
  (
    CASE WHEN frontmatter ->> 'published' IS NOT NULL THEN
      frontmatter ->> 'published'
    ELSE
      ''
    END)::text AS "published_at"
FROM
  page
WHERE
  "file_path" LIKE $1
ORDER BY
  frontmatter ->> 'published' DESC
LIMIT $2 OFFSET $3;

-- name: GetPageByURLPath :one
SELECT
  *
FROM
  page
WHERE
  url_path LIKE @url_path::text;
