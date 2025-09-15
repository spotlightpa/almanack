-- name: UpsertNewsFeedArchives :execrows
WITH raw_json AS (
  SELECT
    jsonb_array_elements(@data::jsonb) AS data
),
feed_items AS (
  SELECT
    data ->> 'id' AS external_id,
    data ->> 'author' AS author,
    CASE WHEN jsonb_typeof(data -> 'authors') = 'array' THEN
      ARRAY (
        SELECT
          jsonb_array_elements_text(data -> 'authors'))
      ELSE
        ARRAY[]::text[]
    END AS authors,
    data ->> 'category' AS category,
    data ->> 'content_html' AS content_html,
    iso_to_timestamptz (data ->> 'date_modified')::timestamptz AS external_updated_at,
    iso_to_timestamptz (data ->> 'date_published')::timestamptz AS external_published_at,
    data ->> 'image' AS image,
    data ->> 'image_credit' AS image_credit,
    data ->> 'image_description' AS image_description,
    data ->> 'language' AS "language",
    data ->> 'title' AS title,
    data ->> 'url' AS url
  FROM
    raw_json)
  INSERT INTO news_feed_item ("external_id", "author", "authors", "category",
    "content_html", "external_updated_at", "external_published_at", "image",
    "image_credit", "image_description", "language", "title", "url")
  SELECT
    "external_id",
    COALESCE("author", ''),
    "authors",
    COALESCE("category", ''),
    COALESCE("content_html", ''),
    "external_updated_at",
    "external_published_at",
    COALESCE("image", ''),
    COALESCE("image_credit", ''),
    COALESCE("image_description", ''),
    COALESCE("language", ''),
    COALESCE("title", ''),
    COALESCE("url", '')
  FROM
    feed_items
  ON CONFLICT ("external_id")
    DO UPDATE SET
      "author" = EXCLUDED.author,
      "authors" = EXCLUDED.authors,
      "category" = EXCLUDED.category,
      "content_html" = EXCLUDED.content_html,
      "external_updated_at" = EXCLUDED.external_updated_at,
      "external_published_at" = EXCLUDED.external_published_at,
      "image" = EXCLUDED.image,
      "image_credit" = EXCLUDED.image_credit,
      "image_description" = EXCLUDED.image_description,
      "language" = EXCLUDED.language,
      "title" = EXCLUDED.title,
      "url" = EXCLUDED.url,
      "uploaded_at" = CASE WHEN news_feed_item.external_updated_at <>
	EXCLUDED.external_updated_at THEN
        NULL
      ELSE
        news_feed_item.uploaded_at
      END,
      "updated_at" = CURRENT_TIMESTAMP;

-- name: ListNewsFeedUpdates :many
SELECT
  *
FROM
  news_feed_item
WHERE
  "uploaded_at" IS NULL;

-- name: UpdateFeedAppleID :one
UPDATE
  news_feed_item
SET
  "apple_id" = $1,
  "apple_share_url" = $2,
  "uploaded_at" = CURRENT_TIMESTAMP,
  "updated_at" = CURRENT_TIMESTAMP
WHERE
  "id" = $3
RETURNING
  *;

-- name: UpdateFeedUploaded :one
UPDATE
  news_feed_item
SET
  "uploaded_at" = CURRENT_TIMESTAMP,
  "updated_at" = CURRENT_TIMESTAMP
WHERE
  "id" = $1
RETURNING
  *;
