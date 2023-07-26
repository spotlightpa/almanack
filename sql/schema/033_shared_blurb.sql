ALTER TABLE "shared_article"
  ADD COLUMN "blurb" text NOT NULL DEFAULT '';

ALTER TABLE "shared_article" DISABLE TRIGGER row_updated_at_on_shared_article_trigger_;

UPDATE
  "shared_article" AS sa
SET
  blurb = gd."metadata" ->> 'blurb'
FROM
  "g_docs_doc" gd
WHERE
  sa."source_type" = 'gdocs'
  AND cast(sa."raw_data" AS bigint) = gd.id;

ALTER TABLE shared_article ENABLE TRIGGER row_updated_at_on_shared_article_trigger_;

---- create above / drop below ----
ALTER TABLE "shared_article"
  DROP COLUMN "blurb";
