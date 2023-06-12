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

-- name: ListImagesByFTS :many
SELECT
  image.*
FROM
  image,
  websearch_to_tsquery('english', @query) tsq
WHERE
  fts @@ tsq
  AND is_uploaded
ORDER BY
  ts_rank(fts, tsq) DESC,
  updated_at DESC
LIMIT $1 OFFSET $2;

-- name: UpsertImage :one
INSERT INTO image ("path", "type", "description", "credit", "keywords",
  "src_url", "is_uploaded")
  VALUES (@path, @type, @description, @credit, @keywords, @src_url, @is_uploaded)
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
    END, keywords = CASE WHEN image.keywords = '' THEN
      excluded.keywords
    ELSE
      image.keywords
    END, src_url = CASE WHEN image.src_url = '' THEN
      excluded.src_url
    ELSE
      image.src_url
    END
  RETURNING
    *;

-- name: UpsertImageWithMD5 :one
INSERT INTO image ("path", "type", "description", "credit", "keywords",
  "src_url", "md5", "bytes", "is_uploaded")
  VALUES (@path, @type, @description, @credit, @keywords, @src_url, @md5, @bytes, TRUE)
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
    END, keywords = CASE WHEN image.keywords = '' THEN
      excluded.keywords
    ELSE
      image.keywords
    END, src_url = CASE WHEN image.src_url = '' THEN
      excluded.src_url
    ELSE
      image.src_url
    END, md5 = CASE WHEN image.md5 = '' THEN
      excluded.md5
    ELSE
      image.md5
    END, bytes = CASE WHEN image.bytes = 0 THEN
      excluded.bytes
    ELSE
      image.bytes
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
  keywords = CASE WHEN @set_keywords::boolean THEN
    @keywords
  ELSE
    keywords
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
  *
FROM
  image
WHERE
  is_uploaded = FALSE
  AND src_url <> '';

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
