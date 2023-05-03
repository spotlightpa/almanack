ALTER TABLE "image"
  ADD COLUMN "md5" bytea NOT NULL DEFAULT '',
  ADD COLUMN "bytes" bigint NOT NULL DEFAULT 0;

CREATE INDEX "image_md5" ON "image" ("md5");

ALTER TABLE "file"
  ADD COLUMN "md5" bytea NOT NULL DEFAULT '',
  ADD COLUMN "bytes" bigint NOT NULL DEFAULT 0;

CREATE INDEX "file_md5" ON "file" ("md5");

CREATE TABLE "image_source" (
  "id" bigserial PRIMARY KEY,
  "image_id" bigint NOT NULL REFERENCES "image" ("id"),
  "url" text NOT NULL UNIQUE
);

INSERT INTO image_source (image_id, url)
SELECT
  id,
  src_url
FROM
  image
WHERE
  src_url <> '';

---- create above / drop below ----
ALTER TABLE "image"
  DROP COLUMN "md5",
  DROP COLUMN "bytes";

ALTER TABLE "file"
  DROP COLUMN "md5",
  DROP COLUMN "bytes";
