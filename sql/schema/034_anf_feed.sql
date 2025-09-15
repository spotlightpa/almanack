CREATE TABLE news_feed_item (
  "id" bigserial PRIMARY KEY,
  "external_id" text NOT NULL UNIQUE,
  "author" text NOT NULL DEFAULT '',
  "authors" text[] NOT NULL DEFAULT '{}',
  "category" text NOT NULL DEFAULT '',
  "content_html" text NOT NULL DEFAULT '',
  "external_updated_at" timestamp with time zone NOT NULL,
  "external_published_at" timestamp with time zone NOT NULL,
  "image" text NOT NULL DEFAULT '',
  "image_credit" text NOT NULL DEFAULT '',
  "image_description" text NOT NULL DEFAULT '',
  "language" text NOT NULL DEFAULT '',
  "title" text NOT NULL DEFAULT '',
  "url" text NOT NULL DEFAULT '',
  "uploaded_at" timestamp with time zone,
  "apple_id" text NOT NULL DEFAULT '',
  "apple_share_url" text NOT NULL DEFAULT '',
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX news_feed_item_published_at_idx ON news_feed_item ("uploaded_at");

---- create above / drop below ----
DROP TABLE news_feed_item;
