BEGIN;
ALTER TABLE article DISABLE TRIGGER row_updated_at_on_article_trigger_;
WITH export_data AS (
  SELECT
    jsonb_array_elements($$ REPLACEME $$::jsonb) AS article_data
),
structured_data AS (
  SELECT
    article_data -> 'spotlightpa_data' ->> 'arc-id' AS arc_id,
    to_timestamp(article_data ->> 'last_published', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz
      AS last_published,
    article_data ->> 'spotlightpa_path' AS spotlightpa_path,
    article_data -> 'spotlightpa_data' || jsonb_build_object('last-arc-sync',
      to_char(transaction_timestamp(), 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS spotlightpa_data
  FROM
    export_data
),
filtered_data AS (
  SELECT
    *
  FROM
    structured_data
  WHERE
    arc_id IS NOT NULL)
UPDATE
  article
SET
  last_published = filtered_data.last_published,
  spotlightpa_data = filtered_data.spotlightpa_data,
  spotlightpa_path = filtered_data.spotlightpa_path
FROM
  filtered_data
WHERE
  article.arc_id = filtered_data.arc_id
  AND article.spotlightpa_path IS NULL
RETURNING
  *;
UPDATE
  article
SET
  spotlightpa_data = spotlightpa_data || jsonb_build_object('internal-id',
    arc_data ->> 'slug')
WHERE
  spotlightpa_data ->> 'internal-id' = ''
  AND spotlightpa_path IS NOT NULL
RETURNING
  *;
WITH export_data AS (
  SELECT
    jsonb_array_elements($$ REPLACEME $$::jsonb) AS article_data
),
structured_data AS (
  SELECT
    article_data -> 'spotlightpa_data' ->> 'arc-id' AS arc_id,
    to_timestamp(article_data ->> 'last_published',
      'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz AS last_published,
    article_data ->> 'spotlightpa_path' AS spotlightpa_path,
    article_data -> 'spotlightpa_data' || jsonb_build_object('last-arc-sync',
      to_char(transaction_timestamp(), 'YYYY-MM-DD"T"HH24:MI:SS"Z"')) AS spotlightpa_data
  FROM
    export_data
),
filtered_data AS (
  SELECT
    *
  FROM
    structured_data
  WHERE
    arc_id IS NULL)
INSERT INTO article (last_published, spotlightpa_path, spotlightpa_data)
SELECT
  last_published,
  spotlightpa_path,
  spotlightpa_data
FROM
  filtered_data
RETURNING
  *;
UPDATE
  article
SET
  spotlightpa_data = spotlightpa_data || jsonb_build_object('internal-id',
    'SPLTKTK')
WHERE
  spotlightpa_data ->> 'internal-id' = ''
  AND spotlightpa_path IS NOT NULL
RETURNING
  *;
ALTER TABLE article ENABLE TRIGGER row_updated_at_on_article_trigger_;
COMMIT;
