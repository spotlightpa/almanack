ALTER TABLE "newsletter"
  ADD COLUMN "description" text NOT NULL DEFAULT '',
  ADD COLUMN "blurb" text NOT NULL DEFAULT '';

---- create above / drop below ----
ALTER TABLE "newsletter"
  DROP COLUMN "description",
  DROP COLUMN "blurb";
