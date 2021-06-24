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
