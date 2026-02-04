-- name: ListActiveAppleNewsChannels :many
SELECT
  *
FROM
  apple_news_channel
WHERE
  active = TRUE
ORDER BY
  name;

-- name: ListAppleNewsChannels :many
SELECT
  *
FROM
  apple_news_channel
ORDER BY
  name;

-- name: GetAppleNewsChannel :one
SELECT
  *
FROM
  apple_news_channel
WHERE
  id = $1;

-- name: CreateAppleNewsChannel :one
INSERT INTO apple_news_channel (name, channel_id, KEY, secret, feed_url, active)
  VALUES (@name, @channel_id, @key, @secret, @feed_url, @active)
RETURNING
  *;

-- name: CreateAppleNewsChannelWithID :one
INSERT INTO apple_news_channel (id, name, channel_id, KEY, secret, feed_url, active)
  VALUES (@id, @name, @channel_id, @key, @secret, @feed_url, @active)
RETURNING
  *;

-- name: UpdateAppleNewsChannel :one
UPDATE
  apple_news_channel
SET
  name = @name,
  channel_id = @channel_id,
  KEY = @key,
  secret = @secret,
  feed_url = @feed_url,
  active = @active,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = @id
RETURNING
  *;

-- name: UpdateAppleNewsChannelLastSynced :exec
UPDATE
  apple_news_channel
SET
  last_synced_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE
  id = $1;

-- name: DeleteAppleNewsChannel :exec
DELETE FROM apple_news_channel
WHERE id = $1;

-- name: UpsertANFChannelItem :one
INSERT INTO anf_channel_item (channel_id, news_feed_item_id, apple_id,
  apple_share_url, uploaded_at)
  VALUES (@channel_id, @news_feed_item_id, @apple_id, @apple_share_url, CURRENT_TIMESTAMP)
ON CONFLICT (channel_id, news_feed_item_id)
  DO UPDATE SET
    apple_id = EXCLUDED.apple_id, apple_share_url = EXCLUDED.apple_share_url,
      uploaded_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
  RETURNING
    *;

-- name: GetANFChannelItem :one
SELECT
  *
FROM
  anf_channel_item
WHERE
  channel_id = $1
  AND news_feed_item_id = $2;

-- name: ListANFChannelItemsForChannel :many
SELECT
  *
FROM
  anf_channel_item
WHERE
  channel_id = $1;

-- name: ListANFChannelItemsNeedingUpload :many
-- Returns news_feed_items that are in the feed but either:
-- 1. Not yet uploaded to this channel (no anf_channel_item row)
-- 2. Updated since last upload (external_updated_at > uploaded_at)
SELECT
  nfi.*
FROM
  news_feed_item nfi
WHERE
  nfi.external_id = ANY (@external_ids::text[])
  AND (NOT EXISTS (
      SELECT
        1
      FROM
        anf_channel_item aci
      WHERE
        aci.channel_id = @channel_id
        AND aci.news_feed_item_id = nfi.id)
      OR EXISTS (
        SELECT
          1
        FROM
          anf_channel_item aci
        WHERE
          aci.channel_id = @channel_id
          AND aci.news_feed_item_id = nfi.id
          AND nfi.external_updated_at > aci.uploaded_at));

-- name: MarkANFChannelItemUploaded :one
UPDATE
  anf_channel_item
SET
  uploaded_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP
WHERE
  channel_id = @channel_id
  AND news_feed_item_id = @news_feed_item_id
RETURNING
  *;
