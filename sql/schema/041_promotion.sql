CREATE TABLE promotion (
  "id" bigserial PRIMARY KEY,
  "name" text NOT NULL DEFAULT '',
  "description" text NOT NULL DEFAULT '',
  "width" bigint NOT NULL DEFAULT 0,
  "height" bigint NOT NULL DEFAULT 0,
  "items" text[] NOT NULL DEFAULT '{}',
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
