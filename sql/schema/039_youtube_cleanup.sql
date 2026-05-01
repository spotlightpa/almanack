ALTER TABLE youtube
  DROP COLUMN "uploaded_at";

---- create above / drop below ----
ALTER TABLE youtube
  ADD COLUMN "uploaded_at" timestamp with time zone;
