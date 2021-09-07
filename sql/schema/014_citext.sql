CREATE EXTENSION IF NOT EXISTS citext;

ALTER TABLE "page"
  ALTER COLUMN "file_path" TYPE citext,
  ALTER COLUMN "url_path" TYPE citext;

ALTER TABLE "domain_roles"
  ALTER COLUMN "domain" TYPE citext;

ALTER TABLE "address_roles"
  ALTER COLUMN "email_address" TYPE citext;

ALTER TABLE "newsletter"
  ALTER COLUMN "spotlightpa_path" TYPE citext;

ALTER TABLE "site_data"
  ALTER COLUMN "key" TYPE citext;

---- create above / drop below ----
ALTER TABLE "page"
  ALTER COLUMN "file_path" TYPE text,
  ALTER COLUMN "url_path" TYPE text;

ALTER TABLE "domain_roles"
  ALTER COLUMN "domain" TYPE text;

ALTER TABLE "address_roles"
  ALTER COLUMN "email_address" TYPE text;

ALTER TABLE "newsletter"
  ALTER COLUMN "spotlightpa_path" TYPE text;

ALTER TABLE "site_data"
  ALTER COLUMN "key" TYPE text;
