ALTER TABLE youtube
  ADD COLUMN "page_id" bigint REFERENCES page (id) UNIQUE,
  ADD COLUMN "description" text NOT NULL DEFAULT '';

---- create above / drop below ----
ALTER TABLE youtube
  DROP COLUMN "page_id";
