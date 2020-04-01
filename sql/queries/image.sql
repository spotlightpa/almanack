-- name: ListImages :many
SELECT
  *
FROM
  image
WHERE
  is_uploaded = TRUE
ORDER BY
  updated_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateImagePlaceholder :execrows
INSERT INTO image ("path", "type")
  VALUES (@path, @type)
ON CONFLICT (path)
  DO NOTHING;

-- name: UpdateImage :one
UPDATE
  image
SET
  credit = CASE WHEN @set_credit::boolean THEN
    @credit
  ELSE
    credit
  END,
  description = CASE WHEN @set_description::boolean THEN
    @description
  ELSE
    description
  END,
  src_url = CASE WHEN src_url = '' THEN
    @src_url
  ELSE
    src_url
  END,
  is_uploaded = TRUE
WHERE
  path = @path
RETURNING
  *;

-- name: GetImageBySourceURL :one
SELECT
  *
FROM
  image
WHERE
  src_url = $1
ORDER BY
  updated_at DESC
LIMIT 1;
