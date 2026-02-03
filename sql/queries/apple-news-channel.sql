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
