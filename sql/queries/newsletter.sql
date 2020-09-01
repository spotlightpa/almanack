-- name: UpdateNewsletterArchives :execrows
WITH raw_json AS (
  SELECT
    jsonb_array_elements(@data::jsonb) AS data
),
campaign AS (
  SELECT
    data ->> 'subject' AS subject,
    data ->> 'archive_url' AS archive_url,
    to_timestamp(data ->> 'published_at'::text,
      -- ISO date
      'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS published_at
  FROM
    raw_json)
  INSERT INTO newsletter ("subject", "archive_url", "published_at", "type")
  SELECT
    subject,
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
