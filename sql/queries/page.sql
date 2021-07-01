-- name: EnsurePage :exec
INSERT INTO page ("path")
  VALUES (@path)
ON CONFLICT (path)
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
  last_published = CASE WHEN @set_last_published::boolean THEN
    CURRENT_TIMESTAMP
  ELSE
    last_published
  END
WHERE
  path = @path
RETURNING
  *;

-- name: GetPage :one
SELECT
  *
FROM
  "page"
WHERE
  path = $1;

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
  "path",
  (frontmatter ->> 'internal-id')::text AS "internal_id",
  (frontmatter ->> 'title')::text AS "title",
  (frontmatter ->> 'description')::text AS "description",
  (frontmatter ->> 'blurb')::text AS "blurb",
  "last_published",
  "created_at",
  "updated_at",
  to_timestamp(frontmatter ->> 'published',
    -- ISO date
    'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz AS "published_at"
FROM
  page
WHERE
  "path" ILIKE $1
ORDER BY
  last_published DESC,
  created_at DESC
LIMIT $2 OFFSET $3;
