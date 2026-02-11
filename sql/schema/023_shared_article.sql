CREATE TABLE arc (
  id bigserial PRIMARY KEY,
  arc_id character varying(26) UNIQUE NOT NULL,
  raw_data jsonb NOT NULL DEFAULT '{}'::jsonb,
  last_updated timestamptz GENERATED ALWAYS AS (iso_to_timestamptz (raw_data
    ->> 'last_updated_date')) STORED,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO arc (arc_id, raw_data, created_at, updated_at)
SELECT
  arc_id,
  arc_data,
  created_at,
  updated_at
FROM
  article
WHERE
  arc_id IS NOT NULL;

CREATE TRIGGER row_updated_at_on_arc_trigger_
  BEFORE UPDATE ON arc
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

CREATE TABLE shared_status (
  id character(1) PRIMARY KEY,
  description text NOT NULL DEFAULT ''
);

INSERT INTO shared_status ("id", "description")
VALUES
  ('U', 'Unshared'),
  ('P', 'Preview'),
  ('S', 'Shared');

CREATE TABLE shared_article (
  id bigserial PRIMARY KEY,
  status character(1) NOT NULL DEFAULT 'U' REFERENCES shared_status (id),
  embargo_until timestamptz,
  note text NOT NULL DEFAULT '',
  "source_type" text NOT NULL DEFAULT '',
  "source_id" text NOT NULL DEFAULT '',
  "raw_data" jsonb NOT NULL DEFAULT '{}'::jsonb,
  page_id bigint REFERENCES page (id) UNIQUE,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX shared_article_unique_source_key ON shared_article
  (source_type, source_id);

INSERT INTO shared_article (status, note, source_type, source_id, raw_data,
  page_id, created_at, updated_at)
SELECT
  CASE WHEN status = 'A' THEN
    'S'
  ELSE
    status
  END,
  note,
  'arc',
  arc_id,
  arc_data,
  (
    SELECT
      page.id
    FROM
      page
    WHERE
      page.file_path = article.spotlightpa_path), --
  created_at, updated_at
FROM
  article
WHERE
  arc_id IS NOT NULL
  AND (status = 'A'
    OR status = 'P');

CREATE TRIGGER row_updated_at_on_shared_article_trigger_
  BEFORE UPDATE ON shared_article
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE arc, shared_article, shared_status;
