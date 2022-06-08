ALTER TABLE "article"
  DROP COLUMN "spotlightpa_data",
  DROP COLUMN "schedule_for",
  DROP COLUMN "last_published";

---- create above / drop below ----
ALTER TABLE "article"
  ADD COLUMN "spotlightpa_data" jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN "schedule_for" timestamptz,
  ADD COLUMN "last_published" timestamptz;
