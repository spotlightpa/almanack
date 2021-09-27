-- name: EnsurePage :exec
INSERT INTO page ("file_path")
  VALUES (@file_path)
ON CONFLICT (file_path)
  DO NOTHING;

-- name: UpdatePage :one
UPDATE
  page
SET
  frontmatter = CASE WHEN @set_frontmatter::boolean THEN
    @frontmatter
  ELSE
    frontmatter
  END,
  body = CASE WHEN @set_body::boolean THEN
    @body
  ELSE
    body
  END,
  schedule_for = CASE WHEN @set_schedule_for::boolean THEN
    @schedule_for
  ELSE
    schedule_for
  END,
  url_path = CASE WHEN @url_path::text != '' THEN
    @url_path
  ELSE
    url_path
  END,
  last_published = CASE WHEN @set_last_published::boolean THEN
    CURRENT_TIMESTAMP
  ELSE
    last_published
  END
WHERE
  file_path = @file_path
RETURNING
  *;

-- name: GetPageByFilePath :one
SELECT
  *
FROM
  "page"
WHERE
  file_path = $1;

-- name: GetPageByID :one
SELECT
  *
FROM
  "page"
WHERE
  id = $1;

-- name: PopScheduledPages :many
UPDATE
  page
SET
  last_published = CURRENT_TIMESTAMP
WHERE
  last_published IS NULL
  AND schedule_for < (CURRENT_TIMESTAMP + '5 minutes'::interval)
RETURNING
  *;

-- ListPages by file path.
-- Cannot use coalesce directly, see https://github.com/kyleconroy/sqlc/issues/780.
-- Treating published_at as text because it sorts faster and we don't do
-- date stuff on the backend, just frontend.
-- name: ListPages :many
WITH paths AS (
  SELECT
    *
  FROM
    page
  WHERE
    "file_path" LIKE $1
),
ordered AS (
  SELECT
    *
  FROM
    paths
  ORDER BY
    frontmatter ->> 'published' DESC
),
selected AS (
  SELECT
    "id",
    "file_path",
    coalesce(frontmatter ->> 'internal-id', '') AS "internal_id",
    coalesce(frontmatter ->> 'title', '') AS "title",
  coalesce(frontmatter ->> 'description', '') AS "description",
  coalesce(frontmatter ->> 'blurb', '') AS "blurb",
  coalesce(frontmatter ->> 'image', '') AS "image",
  coalesce(url_path, '') AS "url_path",
  "last_published",
  "created_at",
  "updated_at",
  "schedule_for",
  coalesce(frontmatter ->> 'published', '') AS "published_at"
FROM
  ordered
)
SELECT
  id,
  file_path::text,
  internal_id::text,
  title::text,
  description::text,
  blurb::text,
  image::text,
  url_path::text,
  last_published,
  created_at,
  updated_at,
  schedule_for,
  published_at::text
FROM
  selected
LIMIT $2 OFFSET $3;

-- name: ListPageIDs :many
SELECT
  "id"
FROM
  page
WHERE
  "file_path" LIKE $1
ORDER BY
  id ASC
LIMIT $2 OFFSET $3;

-- name: GetPageByURLPath :one
SELECT
  *
FROM
  page
WHERE
  url_path ILIKE @url_path::text;

-- name: ListAllTopics :many
WITH topics AS (
  SELECT
    jsonb_array_elements_text(frontmatter -> 'topics') AS topic
  FROM
    page
  WHERE
    frontmatter ->> 'topics' IS NOT NULL
)
SELECT DISTINCT ON (upper(topic)
)
  topic::text
FROM
  topics
ORDER BY
  upper(topic) ASC;

-- name: ListAllSeries :many
WITH series_dates AS (
  SELECT
    jsonb_array_elements_text(frontmatter -> 'series') AS series,
    frontmatter ->> 'published' AS pub_date
  FROM
    page
  WHERE
    frontmatter ->> 'series' IS NOT NULL
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

-- name: ListAllPages :many
WITH ROWS AS (
  SELECT
    id,
    file_path,
    coalesce(frontmatter ->> 'internal-id', '') AS internal_id,
    coalesce(frontmatter ->> 'title', '') AS hed,
    ARRAY (
      SELECT
        jsonb_array_elements_text(
          CASE WHEN frontmatter ->> 'authors' IS NOT NULL THEN
            frontmatter -> 'authors'
          ELSE
            '[]'::jsonb
          END)) AS authors,
      to_timestamp(frontmatter ->> 'published'::text,
        -- ISO date
        'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz AS pub_date
    FROM
      page
    ORDER BY
      pub_date DESC
)
SELECT
  id,
  file_path,
  internal_id::text AS internal_id,
  hed::text AS hed,
  authors::text[] AS authors,
  pub_date::timestamptz AS pub_date
FROM
  ROWS
WHERE
  pub_date IS NOT NULL;

-- name: ListPagesByURLPaths :many
WITH query_paths AS (
  SELECT
    @paths::text[] AS "paths"
),
page_paths AS (
  SELECT
    *
  FROM
    page
  WHERE
    url_path IN (
      SELECT
        unnest("paths")::citext
      FROM
        query_paths)
),
selected AS (
  SELECT
    "file_path",
    coalesce(frontmatter ->> 'internal-id', '') AS "internal_id",
    coalesce(frontmatter ->> 'title', '') AS "title",
  coalesce(frontmatter ->> 'link-title', '') AS "link_title",
  coalesce(frontmatter ->> 'description', '') AS "description",
  coalesce(frontmatter ->> 'blurb', '') AS "blurb",
  coalesce(frontmatter ->> 'image', '') AS "image",
  coalesce(url_path, '') AS "url_path",
  coalesce(frontmatter ->> 'published', '') AS "published_at"
FROM
  page_paths
)
SELECT
  "file_path"::text,
  "internal_id"::text,
  "title"::text,
  "link_title"::text,
  "description"::text,
  "blurb"::text,
  "image"::text,
  "url_path"::text,
  "published_at"::text
FROM
  selected
  CROSS JOIN query_paths
ORDER BY
  array_position(query_paths.paths, selected.url_path::text);
