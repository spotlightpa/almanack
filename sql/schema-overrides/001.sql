-- DO NOT RUN THIS FILE.
-- Hack to remove useless fulltextsearch column from page.
-- See:
-- https://github.com/kyleconroy/sqlc/issues/162
-- https://github.com/kyleconroy/sqlc/issues/1380
SELECT
  fail ();

ALTER TABLE page
  DROP COLUMN fts_doc_en;

ALTER TABLE page
  DROP COLUMN internal_id_fts;

DROP TABLE article;

DROP TABLE article_status;

ALTER TABLE "page"
  DROP COLUMN "published_at";
