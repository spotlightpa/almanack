CREATE TABLE promotion (
  "id" bigserial PRIMARY KEY,
  "name" text NOT NULL DEFAULT '',
  "description" text NOT NULL DEFAULT '',
  "data" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "width" bigint GENERATED ALWAYS AS (COALESCE(("data" ->>
    'width')::bigint, 0)) STORED,
  "height" bigint GENERATED ALWAYS AS (COALESCE(("data" ->>
    'height')::bigint, 0)) STORED,
  "fts_doc_en" tsvector GENERATED ALWAYS AS
    (setweight(to_tsvector('english', "name"), 'A') ||
    setweight(to_tsvector('english', "description"), 'B'))
    STORED,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX promotion_fts_doc_en_idx ON promotion USING gin (fts_doc_en);

---- create above / drop below ----
DROP TABLE promotion;
