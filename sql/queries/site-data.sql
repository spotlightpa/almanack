-- name: GetSiteData :many
SELECT
  *
FROM
  site_data
WHERE
  key = @key::text
  AND published_at IS NULL
  OR published_at = (
    SELECT
      max(published_at) AS max
    FROM
      site_data
    WHERE
      key = @key::text
    GROUP BY
      key)
ORDER BY
  schedule_for ASC;

-- name: PopScheduledSiteChanges :many
UPDATE
  site_data
SET
  published_at = CURRENT_TIMESTAMP
WHERE
  key = @key::text
  AND published_at IS NULL
  AND schedule_for < (CURRENT_TIMESTAMP + '5 minutes'::interval)
RETURNING
  *;

-- name: CleanSiteData :exec
DELETE FROM site_data
WHERE key = @key::text
  AND published_at < (
    SELECT
      max(published_at) AS max
    FROM
      site_data
    WHERE
      key = @key::text
    GROUP BY
      key);

-- name: SetSiteData :exec
UPDATE
  site_data
SET
  "data" = $2
WHERE
  "key" = $1;
