CREATE TABLE newsletter_type (
  shortname text PRIMARY KEY,
  name text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT ''
);

INSERT INTO newsletter_type ("shortname", "name")
  VALUES
    -- 1
    ('investigator', 'The Investigator'),
    -- 2
    ('papost', 'PA Post');

CREATE TABLE newsletter (
  subject text NOT NULL DEFAULT ''::text,
  archive_url text NOT NULL UNIQUE,
  published_at timestamp with time zone NOT NULL,
  type text NOT NULL DEFAULT 'investigator'::text REFERENCES newsletter_type (shortname),
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX newsletter_published_at_type_idx ON newsletter ("published_at"
  timestamptz_ops, "type" text_ops);

CREATE TRIGGER row_updated_at_on_newsletter_trigger_
  BEFORE UPDATE ON newsletter
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE newsletter;

DROP TABLE newsletter_type;
