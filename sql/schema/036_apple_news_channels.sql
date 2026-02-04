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

-- Junction table tracking which items have been uploaded to which channels
CREATE TABLE anf_channel_item (
  "id" bigserial PRIMARY KEY,
  "channel_id" bigint NOT NULL REFERENCES apple_news_channel (id) ON DELETE CASCADE,
  "news_feed_item_id" bigint NOT NULL REFERENCES news_feed_item (id) ON DELETE CASCADE,
  "apple_id" text NOT NULL DEFAULT '',
  "apple_share_url" text NOT NULL DEFAULT '',
  "uploaded_at" timestamp with time zone,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE ("channel_id", "news_feed_item_id")
);

CREATE INDEX anf_channel_item_channel_id_idx ON anf_channel_item ("channel_id");

CREATE INDEX anf_channel_item_news_feed_item_id_idx ON anf_channel_item
  ("news_feed_item_id");

-- Remove upload tracking from news_feed_item (it's now in anf_channel_item)
ALTER TABLE news_feed_item
  DROP COLUMN "uploaded_at",
  DROP COLUMN "apple_id",
  DROP COLUMN "apple_share_url";

---- create above / drop below ----
ALTER TABLE news_feed_item
  ADD COLUMN "uploaded_at" timestamp with time zone,
  ADD COLUMN "apple_id" text NOT NULL DEFAULT '',
  ADD COLUMN "apple_share_url" text NOT NULL DEFAULT '';

DROP TABLE anf_channel_item;

DROP TABLE apple_news_channel;
