ALTER TABLE image
  ADD COLUMN deleted_at timestamptz;

ALTER TABLE file
  ADD COLUMN deleted_at timestamptz;

---- create above / drop below ----
ALTER TABLE image
  DROP COLUMN deleted_at timestamptz;

ALTER TABLE file
  DROP COLUMN deleted_at timestamptz;
