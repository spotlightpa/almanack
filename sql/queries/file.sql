-- name: ListFiles :many
SELECT
  *
FROM
  file
WHERE
  is_uploaded = TRUE
ORDER BY
  updated_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateFilePlaceholder :execrows
INSERT INTO file ("filename", "path", "mime_type")
  VALUES (@filename, @path, @type)
ON CONFLICT (path)
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
  path = @path
RETURNING
  *;
