ALTER TABLE image
  ADD COLUMN keywords text NOT NULL DEFAULT ''::text;

ALTER TABLE image
  ADD COLUMN fts tsvector GENERATED ALWAYS AS
    ((setweight(to_tsvector('english', "description"), 'C')) ||
    (setweight(to_tsvector('english', "credit"), 'B')) ||
    (setweight(to_tsvector('english', "keywords"), 'A')))
    STORED;

CREATE INDEX image_internal_id_fts_idx ON image USING gin (fts);

---- create above / drop below ----
ALTER TABLE image
  DROP COLUMN fts;

ALTER TABLE image
  DROP COLUMN keywords;
