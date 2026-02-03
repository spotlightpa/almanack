CREATE TABLE apple_news_channel (
  "id" bigserial PRIMARY KEY,
  "name" text NOT NULL DEFAULT '',
  "channel_id" text NOT NULL UNIQUE,
  "key" text NOT NULL DEFAULT '',
  "secret" text NOT NULL DEFAULT '',
  "feed_url" text NOT NULL DEFAULT '',
  "active" boolean NOT NULL DEFAULT TRUE,
  "last_synced_at" timestamp with time zone,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add channel reference to news_feed_item
ALTER TABLE news_feed_item
  ADD COLUMN "channel_id" bigint REFERENCES apple_news_channel (id);

-- Update unique constraint to be per-channel
ALTER TABLE news_feed_item
  DROP CONSTRAINT news_feed_item_external_id_key;

ALTER TABLE news_feed_item
  ADD CONSTRAINT news_feed_item_external_id_channel_id_key UNIQUE ("external_id", "channel_id");

---- create above / drop below ----
ALTER TABLE news_feed_item
  DROP CONSTRAINT news_feed_item_external_id_channel_id_key;

ALTER TABLE news_feed_item
  ADD CONSTRAINT news_feed_item_external_id_key UNIQUE ("external_id");

ALTER TABLE news_feed_item
  DROP COLUMN "channel_id";

DROP TABLE apple_news_channel;
