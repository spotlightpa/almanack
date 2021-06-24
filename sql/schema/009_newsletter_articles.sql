ALTER TABLE "newsletter"
  ADD COLUMN "id" BIGSERIAL PRIMARY KEY,
  ADD COLUMN "description" text NOT NULL DEFAULT '',
  ADD COLUMN "blurb" text NOT NULL DEFAULT '',
  ADD COLUMN "spotlightpa_path" text UNIQUE;

---- create above / drop below ----
ALTER TABLE "newsletter"
  DROP COLUMN "id",
  DROP COLUMN "description",
  DROP COLUMN "blurb",
  DROP COLUMN "spotlightpa_path";
