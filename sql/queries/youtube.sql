-- name: UpsertYouTubeFeedArchives :execrows
WITH raw_json AS (
  SELECT
    jsonb_array_elements(@data::jsonb) AS data
),
feed_items AS (
  SELECT
    data ->> 'external_id' AS external_id,
    data ->> 'title' AS title,
    data ->> 'url' AS url,
    data ->> 'thumbnail_url' AS thumbnail_url,
    iso_to_timestamptz (data ->> 'external_updated_at')::timestamptz AS external_updated_at,
    iso_to_timestamptz (data ->> 'external_published_at')::timestamptz AS external_published_at
  FROM
    raw_json)
INSERT INTO youtube ("external_id", "title", "url", "thumbnail_url",
  "external_published_at", "external_updated_at")
SELECT
  "external_id",
  "title",
  "url",
  "thumbnail_url",
  "external_published_at",
  "external_updated_at"
FROM
  feed_items
ON CONFLICT ("external_id")
  DO UPDATE SET
    "external_id" = EXCLUDED.external_id,
    "title" = EXCLUDED.title,
    "url" = EXCLUDED.url,
    "thumbnail_url" = EXCLUDED.thumbnail_url,
    "external_published_at" = EXCLUDED.external_published_at,
    "external_updated_at" = EXCLUDED.external_updated_at,
    "uploaded_at" = CASE WHEN youtube.external_updated_at <>
      EXCLUDED.external_updated_at THEN
      NULL
    ELSE
      youtube.uploaded_at
    END,
    "updated_at" = CURRENT_TIMESTAMP;

-- name: ListYouTubeUpdates :many
SELECT
  *
FROM
  youtube
WHERE
  "uploaded_at" IS NULL;

-- name: UpdateYouTubeUploaded :one
UPDATE
  youtube
SET
  "uploaded_at" = CURRENT_TIMESTAMP,
  "updated_at" = CURRENT_TIMESTAMP
WHERE
  "id" = $1
RETURNING
  *;
