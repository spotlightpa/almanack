ALTER TABLE "shared_article"
  ADD COLUMN "blurb" text NOT NULL DEFAULT '';

ALTER TABLE "shared_article" DISABLE TRIGGER row_updated_at_on_shared_article_trigger_;

UPDATE
  "shared_article"
SET
  "blurb" = "description";

ALTER TABLE shared_article ENABLE TRIGGER row_updated_at_on_shared_article_trigger_;

---- create above / drop below ----
ALTER TABLE "shared_article"
  DROP COLUMN "description";
