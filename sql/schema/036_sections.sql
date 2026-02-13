ALTER TABLE news_feed_item
  ADD COLUMN "topics" text[] NOT NULL DEFAULT '{}';

---- create above / drop below ----
ALTER TABLE news_feed_item
  DROP COLUMN "topics";
