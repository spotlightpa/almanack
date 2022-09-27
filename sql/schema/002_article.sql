CREATE TABLE article_status (
  status_id character (1) PRIMARY KEY,
  description text NOT NULL DEFAULT ''
);

INSERT INTO article_status ("status_id", "description")
  VALUES ('U', 'Unset'),
  ('P', 'Planned'),
  ('A', 'Available');

CREATE TABLE article (
  id serial PRIMARY KEY,
  -- Arc fields
  arc_id character varying(26) UNIQUE,
  arc_data jsonb NOT NULL DEFAULT '{}' ::jsonb,
  -- SpotlightPA.org fields
  spotlightpa_path text UNIQUE,
  spotlightpa_data jsonb NOT NULL DEFAULT '{}' ::jsonb,
  schedule_for timestamp with time zone,
  last_published timestamp with time zone,
  -- Almanack fields
  note text NOT NULL DEFAULT '',
  status character (1) NOT NULL DEFAULT 'U' REFERENCES
    article_status (status_id),
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_article_trigger_
  BEFORE UPDATE ON article
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE article;

DROP TABLE article_status;
