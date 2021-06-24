-- name: UpdateNewsletterArchives :execrows
WITH raw_json AS (
  SELECT
    jsonb_array_elements(@data::jsonb) AS data
),
campaign AS (
  SELECT
    data ->> 'subject' AS subject,
    data ->> 'blurb' AS blurb,
    data ->> 'description' AS description,
    data ->> 'archive_url' AS archive_url,
    to_timestamp(data ->> 'published_at'::text,
      -- ISO date
      'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS published_at
  FROM
    raw_json)
  INSERT INTO newsletter ("subject", "blurb", "description", "archive_url",
    "published_at", "type")
  SELECT
    subject,
    blurb,
    description,
    archive_url,
    published_at,
    @type
  FROM
    campaign
  ON CONFLICT
    DO NOTHING;

-- name: ListNewsletters :many
SELECT
  *
FROM
  newsletter
WHERE
  "type" = $1
ORDER BY
  published_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUnpublishedNewsletters :many
SELECT
  *
FROM
  newsletter
WHERE
  "spotlightpa_path" IS NULL
ORDER BY
  published_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateNewsletter :one
UPDATE
  newsletter
SET
  "spotlightpa_path" = $2
WHERE
  id = $1
RETURNING
  *;
