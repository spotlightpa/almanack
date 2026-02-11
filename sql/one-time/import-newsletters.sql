WITH raw_json AS (
  SELECT
    $$ $$::jsonb AS data
),
campaign_json AS (
  SELECT
    jsonb_array_elements(data -> 'campaigns') AS campaign
  FROM
    raw_json
),
campaign AS (
  SELECT
    to_timestamp(campaign ->> 'send_time', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS published_at,
    campaign ->> 'archive_url' AS archive_url,
    campaign -> 'settings' ->> 'subject_line' AS subject
  FROM
    campaign_json
),
filtered_campaign AS (
  SELECT
    *
  FROM
    campaign
  WHERE
    campaign.archive_url NOT IN (
      SELECT
        archive_url
      FROM
        newsletter))
INSERT INTO newsletter (subject, archive_url, published_at)
SELECT
  subject,
  archive_url,
  published_at
FROM
  filtered_campaign
RETURNING
  *;
