CREATE OR REPLACE FUNCTION iso_to_timestamptz (text)
  RETURNS timestamptz
  AS $$
  SELECT
    CASE WHEN $1 ~ '\d{4}-\d\d-\d\dT\d\d:\d\d' THEN
      $1::timestamptz
    ELSE
      NULL
    END
$$
LANGUAGE sql
IMMUTABLE;

ALTER TABLE page
  ADD COLUMN published_at timestamptz GENERATED ALWAYS AS (iso_to_timestamptz
    (frontmatter ->> 'published')) STORED;

ALTER TABLE page
  ADD COLUMN fts_doc_en tsvector GENERATED ALWAYS AS
    (setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'title', '')), 'A') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'linktitle', '')), 'A') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'title-tag', '')), 'A') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'og-title', '')), 'A') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'twitter-title', '')), 'A') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'description', '')), 'B') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'blurb', '')), 'B') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'byline', '')), 'C') --
  || setweight(to_tsvector('english', --
  coalesce(frontmatter ->> 'kicker', '')), 'C') --
  || setweight(to_tsvector('english', body), 'D')) STORED;

CREATE INDEX page_fts_doc_en_idx ON page USING gin (fts_doc_en);

ALTER TABLE page
  ADD COLUMN internal_id_fts tsvector GENERATED ALWAYS AS
    (to_tsvector('simple', coalesce(frontmatter ->> 'internal-id',
    ''))) STORED;

CREATE INDEX page_internal_id_fts_idx ON page USING gin (internal_id_fts);

---- create above / drop below ----
ALTER TABLE page
  DROP COLUMN published_at,
  DROP COLUMN fts_doc_en,
  DROP COLUMN internal_id_fts;
