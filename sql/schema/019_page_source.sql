ALTER TABLE "page"
  ADD COLUMN "source_type" text NOT NULL DEFAULT '',
  ADD COLUMN "source_id" text NOT NULL DEFAULT '';

UPDATE
  page
SET
  source_type = 'arc',
  source_id = frontmatter ->> 'arc-id'
WHERE
  frontmatter ->> 'arc-id' IS NOT NULL;

WITH nl_to_page AS (
  SELECT
    newsletter.id AS nl_id,
    page.id AS p_id
  FROM
    newsletter
    LEFT JOIN page ON newsletter.spotlightpa_path = page.file_path
  WHERE
    newsletter.spotlightpa_path != '')
UPDATE
  page
SET
  source_type = 'mailchimp',
  source_id = nl_id
FROM
  nl_to_page
WHERE
  id = p_id;

UPDATE
  page
SET
  source_type = 'manual'
WHERE
  source_type = '';

---- create above / drop below ----
ALTER TABLE "page"
  DROP COLUMN "source_type",
  DROP COLUMN "source_id";
