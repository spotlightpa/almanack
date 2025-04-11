-- name: CreatePageV2 :one
INSERT INTO page ("file_path", "source_type", "source_id")
  VALUES (@file_path, @source_type, @source_id)
RETURNING
  *;

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
    publication_date DESC
)
SELECT
  id,
  file_path::text,
  coalesce(frontmatter ->> 'internal-id', '')::text AS "internal_id",
  coalesce(frontmatter ->> 'title', '')::text AS "title",
  coalesce(frontmatter ->> 'description', '')::text AS "description",
  coalesce(frontmatter ->> 'blurb', '')::text AS "blurb",
  coalesce(frontmatter ->> 'image', '')::text AS "image",
  coalesce(url_path, '')::text AS "url_path",
  last_published,
  created_at,
  updated_at,
  schedule_for,
  publication_date
FROM
  ordered
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
    json_array_elements_text( --
      coalesce((frontmatter ->> 'topics'), '[]')::json) AS topic
  FROM
    page
)
SELECT DISTINCT ON (upper(topic)
)
  topic::text
FROM
  topics
ORDER BY
  upper(topic) ASC;

-- name: ListAllSeries :many
WITH page_series AS (
  SELECT
    json_array_elements_text( --
      coalesce((frontmatter ->> 'series'), '[]')::json) AS series,
    publication_date
  FROM
    page
),
series_dates AS (
  SELECT
    *
  FROM
    page_series
  ORDER BY
    publication_date DESC,
    series DESC
),
distinct_series_dates AS (
  SELECT DISTINCT ON (series)
    *
  FROM
    series_dates
  ORDER BY
    series DESC,
    publication_date DESC
)
SELECT
  series::text
FROM
  distinct_series_dates
ORDER BY
  publication_date DESC;

-- name: ListAllPages :many
SELECT
  id,
  file_path,
  coalesce(frontmatter ->> 'internal-id', '')::text AS internal_id,
  coalesce(frontmatter ->> 'title', '')::text AS hed,
  ARRAY (
    SELECT
      jsonb_array_elements_text(
        CASE WHEN frontmatter ->> 'authors' IS NOT NULL THEN
          frontmatter -> 'authors'
        ELSE
          '[]'::jsonb
        END))::text[] AS authors,
  publication_date::timestamptz AS pub_date
FROM
  page
WHERE
  publication_date IS NOT NULL
ORDER BY
  publication_date DESC;

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
        query_paths))
SELECT
  "file_path"::text,
  coalesce(frontmatter ->> 'internal-id', '')::text AS "internal_id",
  coalesce(frontmatter ->> 'title', '')::text AS "title",
  coalesce(frontmatter ->> 'link-title', '')::text AS "link_title",
  coalesce(frontmatter ->> 'description', '')::text AS "description",
  coalesce(frontmatter ->> 'blurb', '')::text AS "blurb",
  coalesce(frontmatter ->> 'image', '')::text AS "image",
  coalesce(url_path, '')::text AS "url_path",
  publication_date::timestamptz
FROM
  page_paths
  CROSS JOIN query_paths
ORDER BY
  array_position(query_paths.paths, url_path::text);

-- name: UpdatePageRawContent :one
UPDATE
  page
SET
  frontmatter = frontmatter || jsonb_build_object('raw-content', @raw_content::text)
WHERE
  id = @id
RETURNING
  *;

-- name: ListPagesByFTS :many
WITH query AS (
  SELECT
    id,
    ts_rank(fts_doc_en, tsq) AS rank
  FROM
    page,
    websearch_to_tsquery('english', @query::text) tsq
  WHERE
    fts_doc_en @@ tsq
  ORDER BY
    rank DESC
  LIMIT $1
)
SELECT
  page.*
FROM
  page
  JOIN query USING (id)
ORDER BY
  publication_date DESC;

-- name: ListPagesByPublished :many
SELECT
  *
FROM
  page
ORDER BY
  publication_date DESC
LIMIT $1 OFFSET $2;

-- name: ListPagesByInternalID :many
WITH query AS (
  SELECT
    id,
    ts_rank(internal_id_fts, id_tsq) AS rank
  FROM
    page,
    tsquery (lower(@query::text)) id_tsq
  WHERE
    internal_id_fts @@ id_tsq
  ORDER BY
    rank DESC
  LIMIT $1
)
SELECT
  page.*
FROM
  page
  JOIN query USING (id)
ORDER BY
  publication_date DESC;
