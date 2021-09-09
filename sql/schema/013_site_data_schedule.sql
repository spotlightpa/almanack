ALTER TABLE "site_data"
  ALTER COLUMN "id" TYPE bigint,
  DROP CONSTRAINT "site_data_key_key",
  ADD COLUMN "schedule_for" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ADD COLUMN "published_at" timestamptz,
  ADD CONSTRAINT "site_data_key_schedule_for_key" UNIQUE ("key", "schedule_for");

UPDATE
  "site_data"
SET
  "published_at" = CURRENT_TIMESTAMP
WHERE
  "published_at" IS NULL;

---- create above / drop below ----
DELETE FROM "site_data"
WHERE "published_at" IS NULL;

ALTER TABLE "site_data"
  ALTER COLUMN "id" TYPE integer,
  DROP CONSTRAINT "site_data_key_schedule_for_key",
  ADD CONSTRAINT "site_data_key_key" UNIQUE ("key"),
  DROP COLUMN "schedule_for",
  DROP COLUMN "published_at";
