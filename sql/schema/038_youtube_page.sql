ALTER TABLE youtube
  ADD COLUMN "page_id" bigint REFERENCES page (id) UNIQUE;

---- create above / drop below ----
ALTER TABLE youtube
  DROP COLUMN "page_id";
