ALTER TABLE "newsletter"
  ADD COLUMN "description" text NOT NULL DEFAULT '',
  ADD COLUMN "blurb" text NOT NULL DEFAULT '',
  ADD COLUMN "spotlightpa_path" text UNIQUE;

CREATE TABLE page (
  id bigserial PRIMARY KEY,
  path text NOT NULL UNIQUE,
  frontmatter jsonb NOT NULL DEFAULT '{}' ::jsonb,
  body text NOT NULL DEFAULT '',
  schedule_for timestamp with time zone,
  last_published timestamp with time zone,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

---- create above / drop below ----
ALTER TABLE "newsletter"
  DROP COLUMN "description",
  DROP COLUMN "blurb",
  DROP COLUMN "spotlightpa_path";

DROP TABLE "page";
