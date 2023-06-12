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

ALTER TABLE shared_article
  DROP COLUMN lede_image_source;

ALTER TABLE image
  DROP COLUMN fts;
