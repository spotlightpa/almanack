ALTER TABLE image
  ADD COLUMN "is_licensed" boolean NOT NULL DEFAULT TRUE;

ALTER TABLE image DISABLE TRIGGER row_updated_at_on_image_trigger_;

UPDATE
  image
SET
  "is_licensed" = FALSE
WHERE
  image.fts @@ websearch_to_tsquery('english', 'inquirer');

ALTER TABLE image ENABLE TRIGGER row_updated_at_on_image_trigger_;

---- create above / drop below ----
ALTER TABLE image
  DROP COLUMN "is_licensed";
