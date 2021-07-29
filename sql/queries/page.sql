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

-- name: ListPages :many
SELECT
  "id",
  "file_path",
  (frontmatter ->> 'internal-id')::text AS "internal_id",
  (frontmatter ->> 'title')::text AS "title",
  (frontmatter ->> 'description')::text AS "description",
  (frontmatter ->> 'blurb')::text AS "blurb",
  coalesce("url_path", ''),
  "last_published",
  "created_at",
  "updated_at",
  "schedule_for",
  to_timestamp(frontmatter ->> 'published',
    -- ISO date
    'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz AS "published_at"
FROM
  page
WHERE
  "file_path" ILIKE $1
ORDER BY
  published_at DESC
LIMIT $2 OFFSET $3;
