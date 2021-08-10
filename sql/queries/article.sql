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
    article.status = 'U';

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
  article AS "new"
SET
  spotlightpa_data = @spotlightpa_data,
  schedule_for = @schedule_for,
  spotlightpa_path = CASE WHEN "new".spotlightpa_path IS NULL THEN
    @spotlightpa_path
  ELSE
    "new".spotlightpa_path
  END
FROM
  article AS "old"
WHERE
  "new".id = "old".id
  AND "old".arc_id = @arc_id
RETURNING
  "old".schedule_for;

-- name: UpdateSpotlightPAArticleLastPublished :one
UPDATE
  article AS "new"
SET
  last_published = CURRENT_TIMESTAMP
FROM
  article AS "old"
WHERE
  "new".id = "old".id
  AND "old".arc_id = @arc_id::text
RETURNING
  "old".last_published;

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

-- name: GetArticleByArcID :one
SELECT
  *
FROM
  article
WHERE
  arc_id = @arc_id::text;

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
  arc_data ->> 'last_updated_date' DESC;

-- name: ListAllArcArticles :many
SELECT
  *
FROM
  article
WHERE
  arc_id IS NOT NULL
ORDER BY
  arc_data ->> 'last_updated_date' DESC
LIMIT $1 OFFSET $2;

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
  arc_data ->> 'last_updated_date' DESC
LIMIT $1 OFFSET $2;

-- name: ListSpotlightPAArticles :many
SELECT
  id,
  coalesce(arc_id, '')::text AS arc_id,
  spotlightpa_path::text,
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

-- name: ListAllTopics :many
WITH topic_dates AS (
  SELECT
    jsonb_array_elements_text(spotlightpa_data -> 'topics') AS topic,
    spotlightpa_data ->> 'pub-date' AS pub_date
  FROM
    article
  WHERE
    spotlightpa_data -> 'topics' IS NOT NULL
  ORDER BY
    pub_date DESC,
    topic DESC
),
distinct_topic_dates AS (
  SELECT DISTINCT ON (topic)
    *
  FROM
    topic_dates
  ORDER BY
    topic DESC,
    pub_date DESC
)
SELECT
  topic::text
FROM
  distinct_topic_dates
ORDER BY
  pub_date DESC;

-- name: ListAllSeries :many
WITH series_dates AS (
  SELECT
    jsonb_array_elements_text(spotlightpa_data -> 'series') AS series,
    spotlightpa_data ->> 'pub-date' AS pub_date
  FROM
    article
  WHERE
    spotlightpa_data -> 'series' IS NOT NULL
  ORDER BY
    pub_date DESC,
    series DESC
),
distinct_series_dates AS (
  SELECT DISTINCT ON (series)
    *
  FROM
    series_dates
  ORDER BY
    series DESC,
    pub_date DESC
)
SELECT
  series::text
FROM
  distinct_series_dates
ORDER BY
  pub_date DESC;
