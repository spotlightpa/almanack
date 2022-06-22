ALTER TABLE "file"
  ALTER COLUMN "id" TYPE bigint;

ALTER TABLE "image"
  ALTER COLUMN "id" TYPE bigint;

ALTER TABLE "domain_roles"
  ALTER COLUMN "id" TYPE bigint;

ALTER TABLE "address_roles"
  ALTER COLUMN "id" TYPE bigint;

---- create above / drop below ----
ALTER TABLE "file"
  ALTER COLUMN "id" TYPE integer;

ALTER TABLE "image"
  ALTER COLUMN "id" TYPE integer;

ALTER TABLE "domain_roles"
  ALTER COLUMN "id" TYPE integer;

ALTER TABLE "address_roles"
  ALTER COLUMN "id" TYPE integer;
