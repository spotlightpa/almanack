ALTER TABLE image
  ADD COLUMN width int NOT NULL DEFAULT 0,
  ADD COLUMN height int NOT NULL DEFAULT 0;

---- create above / drop below ----
ALTER TABLE image
  DROP COLUMN width,
  DROP COLUMN height;
