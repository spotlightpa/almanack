CREATE TABLE apple_news_channel (
  "id" bigserial PRIMARY KEY,
  "name" text NOT NULL DEFAULT '',
  "apple_channel_id" text NOT NULL UNIQUE,
  "key" text NOT NULL DEFAULT '',
  "secret" text NOT NULL DEFAULT '',
  "source_feed_url" text NOT NULL DEFAULT '',
  "last_synced_at" timestamp with time zone,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Give existing items the right source_feed_url
ALTER TABLE news_feed_item
  ADD COLUMN "source_feed_url" text NOT NULL DEFAULT 'https://www.spotlightpa.org/feeds/full.json';

ALTER TABLE news_feed_item
  ALTER COLUMN "source_feed_url" SET DEFAULT '';

ALTER TABLE news_feed_item
  DROP CONSTRAINT news_feed_item_external_id_key;

ALTER TABLE news_feed_item
  ADD CONSTRAINT news_feed_item_external_id_source_feed_url_key UNIQUE
    ("external_id", "source_feed_url");

---- create above / drop below ----
DROP TABLE apple_news_channel;

ALTER TABLE news_feed_item
  DROP CONSTRAINT news_feed_item_external_id_source_feed_url_key;

ALTER TABLE news_feed_item
  DROP COLUMN "source_feed_url";

ALTER TABLE news_feed_item
  ADD CONSTRAINT news_feed_item_external_id_key UNIQUE ("external_id");
