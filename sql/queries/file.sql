-- name: ListFiles :many
SELECT
  *
FROM
  file
WHERE
  is_uploaded = TRUE
  AND deleted_at IS NULL
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

-- name: ListFilesWhereNoMD5 :many
SELECT
  *
FROM
  file
WHERE
  md5 = ''
  AND is_uploaded
  AND deleted_at IS NULL
ORDER BY
  created_at ASC
LIMIT $1;

-- name: UpdateFileMD5Size :one
UPDATE
  file
SET
  md5 = @md5,
  bytes = @bytes
WHERE
  id = @id
RETURNING
  *;
