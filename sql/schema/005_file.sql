CREATE TABLE file (
  id serial PRIMARY KEY,
  path text NOT NULL UNIQUE,
  filename text NOT NULL DEFAULT '',
  mime_type text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',
  is_uploaded boolean NOT NULL DEFAULT FALSE,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_file_trigger_
  BEFORE UPDATE ON file
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE file;
