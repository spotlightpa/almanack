-- name: ListFiles :many
SELECT
  *
FROM
  file
WHERE
  is_uploaded = TRUE
ORDER BY
  created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateFilePlaceholder :execrows
INSERT INTO file ("filename", "url", "mime_type")
  VALUES (@filename, @url, @type)
ON CONFLICT (url)
  DO NOTHING;

-- name: UpdateFile :one
UPDATE
  file
SET
  description = CASE WHEN @set_description::boolean THEN
    @description
  ELSE
    description
  END,
  is_uploaded = TRUE
WHERE
  url = @url
RETURNING
  *;
