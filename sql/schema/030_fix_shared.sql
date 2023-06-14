ALTER TABLE shared_article
  DROP COLUMN lede_image_source;

---- create above / drop below ----
ALTER TABLE shared_article
  ADD COLUMN "lede_image_source" text NOT NULL DEFAULT '',;
