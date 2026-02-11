DROP TABLE "article";

DROP TABLE "article_status";

ALTER TABLE "page"
  DROP COLUMN "published_at";

ALTER TABLE "shared_article" RENAME COLUMN "lede_image_source" TO "lede_image_credit";

ALTER TABLE "shared_article"
  ADD COLUMN "lede_image_source" text NOT NULL DEFAULT '';

CREATE INDEX "shared_article_publication_date_index" ON "shared_article"
  ("publication_date");

---- create above / drop below ----
CREATE TABLE article_status (
  status_id character(1) PRIMARY KEY,
  description text NOT NULL DEFAULT ''::text
);

INSERT INTO article_status ("status_id", "description")
VALUES
  ('U', 'Unset'),
  ('P', 'Planned'),
  ('A', 'Available');

CREATE TABLE article (
  id serial PRIMARY KEY,
  arc_id character varying(26) UNIQUE,
  arc_data jsonb NOT NULL DEFAULT '{}'::jsonb,
  spotlightpa_path text UNIQUE,
  note text NOT NULL DEFAULT ''::text,
  status character(1) NOT NULL DEFAULT 'U'::bpchar REFERENCES
    article_status (status_id),
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX "shared_article_publication_date_index";

ALTER TABLE "shared_article"
  DROP COLUMN "lede_image_source";

ALTER TABLE "shared_article" RENAME COLUMN "lede_image_credit" TO "lede_image_source";

ALTER TABLE "page"
  ADD COLUMN published_at timestamp with time zone GENERATED ALWAYS AS
    (iso_to_timestamptz (frontmatter ->> 'published'::text)) STORED;
