-- name: UpsertYouTubeFeedArchives :execrows
WITH raw_json AS (
  SELECT
    jsonb_array_elements(@data::jsonb) AS data
),
feed_items AS (
  SELECT
    data ->> 'external_id' AS "external_id",
    data ->> 'title' AS "title",
    data ->> 'url' AS url,
    data ->> 'thumbnail_url' AS "thumbnail_url",
    iso_to_timestamptz (data ->> 'external_updated_at')::timestamptz AS "external_updated_at",
    iso_to_timestamptz (data ->> 'external_published_at')::timestamptz AS "external_published_at"
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
    "title" = EXCLUDED.title,
    "url" = EXCLUDED.url,
    "thumbnail_url" = EXCLUDED.thumbnail_url,
    "external_published_at" = EXCLUDED.external_published_at,
    "external_updated_at" = EXCLUDED.external_updated_at,
    "updated_at" = CURRENT_TIMESTAMP;

-- name: ResetYouTubeMaxID :exec
SELECT
  setval('youtube_id_seq', (
      SELECT
        MAX(id)
      FROM youtube));

-- name: ListYouTubeWhereNoPage :many
SELECT
  *
FROM
  youtube
WHERE
  "page_id" IS NULL;

-- name: ListYouTubeWhereShort :many
SELECT
  *
FROM
  youtube
WHERE
  "url" LIKE '%/shorts/%'
ORDER BY
  "external_published_at" DESC
LIMIT $1 OFFSET $2;

-- name: ListYouTubeWhereRegular :many
SELECT
  *
FROM
  youtube
WHERE
  "url" LIKE '%/watch?v=%'
ORDER BY
  "external_published_at" DESC
LIMIT $1 OFFSET $2;

-- name: UpdateYouTubePage :one
UPDATE
  youtube
SET
  "page_id" = $1,
  "updated_at" = CURRENT_TIMESTAMP
WHERE
  "id" = $2
RETURNING
  *;
