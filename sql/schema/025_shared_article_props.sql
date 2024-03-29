ALTER TABLE "shared_article"
  ADD COLUMN "publication_date" timestamptz,
  ADD COLUMN "internal_id" text NOT NULL DEFAULT '',
  ADD COLUMN "byline" text NOT NULL DEFAULT '',
  ADD COLUMN "budget" text NOT NULL DEFAULT '',
  ADD COLUMN "hed" text NOT NULL DEFAULT '',
  ADD COLUMN "description" text NOT NULL DEFAULT '',
  ADD COLUMN "lede_image" text NOT NULL DEFAULT '',
  ADD COLUMN "lede_image_source" text NOT NULL DEFAULT '',
  ADD COLUMN "lede_image_description" text NOT NULL DEFAULT '',
  ADD COLUMN "lede_image_caption" text NOT NULL DEFAULT '';

UPDATE
  "shared_article"
SET
  "internal_id" = raw_data ->> 'slug',
  "budget" = raw_data -> 'planning' ->> 'budget_line',
  "hed" = raw_data -> 'headlines' ->> 'basic',
  "description" = raw_data -> 'description' ->> 'basic',
  "publication_date" = iso_to_timestamptz ( --
    raw_data -> 'planning' -> 'scheduling' ->> 'planned_publish_date')
WHERE
  "source_type" = 'arc';

---- create above / drop below ----
ALTER TABLE "shared_article"
  DROP COLUMN "publication_date",
  DROP COLUMN "internal_id",
  DROP COLUMN "byline",
  DROP COLUMN "budget",
  DROP COLUMN "lede_image",
  DROP COLUMN "lede_image_source",
  DROP COLUMN "lede_image_description",
  DROP COLUMN "lede_image_caption",
  DROP COLUMN "hed",
  DROP COLUMN "description";
