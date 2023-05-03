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

-- name: UpsertImageSource :exec
INSERT INTO image_source (image_id, url)
  VALUES (@image_id, @url)
ON CONFLICT (url)
  DO NOTHING;

-- name: InsertImagePlaceholder :exec
WITH i_id AS (
INSERT INTO image ("path", "type", "description", "credit")
    VALUES (@path, @type, @description, @credit)
  RETURNING
    *)
  INSERT INTO image_source (image_id, url)
  SELECT
    i_id.id,
    @source_url
  FROM
    i_id
  RETURNING
    *;

-- name: UpsertImage :one
INSERT INTO image ("path", "type", "description", "credit", "is_uploaded")
  VALUES (@path, @type, @description, @credit, TRUE)
ON CONFLICT (path)
  DO UPDATE SET
    credit = CASE WHEN image.credit = '' THEN
      excluded.credit
    ELSE
      image.credit
    END, description = CASE WHEN image.description = '' THEN
      excluded.description
    ELSE
      image.description
    END
  RETURNING
    *;

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
  is_uploaded = TRUE
WHERE
  path = @path
RETURNING
  *;

-- name: GetImageBySourceURL :one
SELECT
  image.*
FROM
  image
  LEFT JOIN image_source ON (image.id = image_source.image_id)
WHERE
  image_source.url = $1;

-- name: GetImageByMD5 :one
SELECT
  *
FROM
  image
WHERE
  md5 = $1
ORDER BY
  created_at DESC
LIMIT 1;

-- ListImageWhereNotUploaded has no limit
-- because we want them all uploaded,
-- but revisit if queue gets too long.
-- name: ListImageWhereNotUploaded :many
SELECT
  "path",
  "url"::text
FROM
  image
  LEFT JOIN image_source ON (image.id = image_source.image_id)
WHERE
  is_uploaded = FALSE
  AND url IS NOT NULL;

-- name: GetImageTypeForExtension :one
SELECT
  *
FROM
  image_type
WHERE
  @extension::text = ANY (extensions);

-- name: GetImageByPath :one
SELECT
  *
FROM
  "image"
WHERE
  "path" = $1;

-- name: ListImagesWhereNoMD5 :many
SELECT
  *
FROM
  image
WHERE
  md5 = ''
  AND is_uploaded
ORDER BY
  created_at ASC
LIMIT $1;

-- name: UpdateImageMD5Size :one
UPDATE
  image
SET
  md5 = @md5,
  bytes = @bytes
WHERE
  id = @id
RETURNING
  *;
