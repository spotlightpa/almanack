-- name: UpdateArcArticles :exec
WITH arc_table AS (
  SELECT
    jsonb_array_elements(@arc_items::jsonb) AS article_data)
INSERT INTO article (arc_id, arc_data)
SELECT
  article_data ->> '_id',
  article_data
FROM
  arc_table
ON CONFLICT (arc_id)
  DO UPDATE SET
    arc_data = excluded.arc_data
  WHERE
    article.status <> 'A';

-- name: UpdateAlmanackArticle :one
UPDATE
  article
SET
  status = @status,
  note = @note,
  arc_data = CASE WHEN @set_arc_data::bool THEN
    @arc_data::jsonb
  ELSE
    arc_data
  END
WHERE
  arc_id = @arc_id
RETURNING
  *;

-- name: UpdateSpotlightPAArticle :one
UPDATE
  article
SET
  spotlightpa_data = @spotlightpa_data,
  schedule_for = @schedule_for,
  spotlightpa_path = CASE WHEN spotlightpa_path IS NULL THEN
    @spotlightpa_path
  ELSE
    spotlightpa_path
  END,
  last_published = CASE WHEN @set_last_published::bool THEN
    CURRENT_TIMESTAMP
  ELSE
    last_published
  END
WHERE
  arc_id = @arc_id
RETURNING
  *;

-- name: PopScheduled :many
UPDATE
  article
SET
  last_published = CURRENT_TIMESTAMP
WHERE
  last_published IS NULL
  AND schedule_for < (CURRENT_TIMESTAMP + '5 minutes'::interval)
RETURNING
  *;

-- name: GetArticle :one
SELECT
  *
FROM
  article
WHERE
  arc_id = $1;

-- name: GetArticleByDBID :one
SELECT
  *
FROM
  article
WHERE
  id = $1;

-- name: ListUpcoming :many
SELECT
  *
FROM
  article
ORDER BY
  arc_data -> 'last_updated_date' DESC;

-- name: ListAllArticles :many
SELECT
  *
FROM
  article
ORDER BY
  arc_data -> 'last_updated_date' DESC;

-- name: ListAvailableArticles :many
SELECT
  *
FROM
  article
WHERE
  status <> 'U'
ORDER BY
  CASE status
  WHEN 'P' THEN
    '0'
  WHEN 'A' THEN
    '1'
  END ASC,
  arc_data -> 'last_updated_date' DESC;

-- name: ListSpotlightPAArticles :many
SELECT
  id,
  arc_id::text,
  (spotlightpa_data ->> 'internal-id')::text AS internal_id,
  (spotlightpa_data ->> 'hed')::text AS hed,
  ARRAY (
    SELECT
      jsonb_array_elements_text(spotlightpa_data -> 'authors'))::text[] AS authors,
  to_timestamp(spotlightpa_data ->> 'pub-date'::text,
    -- ISO date
    'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS pub_date
FROM
  article
WHERE
  spotlightpa_path IS NOT NULL
ORDER BY
  pub_date DESC;

-- name: GetArticleIDFromSlug :one
SELECT
  arc_id::text
FROM ( SELECT DISTINCT ON (slug)
    arc_id,
    spotlightpa_data ->> 'slug' AS slug,
    created_at
  FROM
    article
  ORDER BY
    slug,
    created_at DESC) AS t
WHERE
  slug = @slug::text;
