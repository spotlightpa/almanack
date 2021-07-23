ALTER TABLE "page" RENAME COLUMN "path" TO "file_path";

ALTER TABLE "page"
  ADD COLUMN "url_path" text,
  ADD UNIQUE ("url_path");

---- create above / drop below ----
ALTER TABLE "page" RENAME COLUMN "file_path" TO "path";

ALTER TABLE "page"
  DROP COLUMN "url_path";
