-- Fix bug with image-caption v. description
UPDATE
  article
SET
  spotlightpa_data = spotlightpa_data || jsonb_build_object('image-description',
    spotlightpa_data ->> 'image-caption')
WHERE
  spotlightpa_data ->> 'image-caption' IS NOT NULL
RETURNING
  *;

CREATE TABLE image_type (
  name text PRIMARY KEY,
  mime text NOT NULL,
  extensions text[]
);

INSERT INTO image_type (name, mime, extensions)
  VALUES
    --
    ('jpeg', 'image/jpeg', '{jpg,jpeg}'),
    --
    ('png', 'image/png', '{png}');

CREATE TABLE image (
  id serial PRIMARY KEY,
  path text NOT NULL UNIQUE,
  type text NOT NULL DEFAULT 'jpeg' ::text REFERENCES image_type (name),
  description text NOT NULL DEFAULT '',
  credit text NOT NULL DEFAULT '',
  src_url text NOT NULL DEFAULT '',
  is_uploaded boolean NOT NULL DEFAULT FALSE,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_image_trigger_
  BEFORE UPDATE ON image
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE image;

DROP TABLE image_type;
