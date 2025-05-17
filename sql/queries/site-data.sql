-- name: GetSiteData :many
SELECT
  *
FROM
  site_data
WHERE
  key = @key::text
  AND (published_at IS NULL
    OR published_at = (
      SELECT
        max(published_at) AS max
      FROM
        site_data
      WHERE
        key = @key::text
      GROUP BY
        key))
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

-- name: UpsertSiteData :exec
INSERT INTO site_data ("key", "data", "schedule_for")
  VALUES (@key, @data, @schedule_for)
ON CONFLICT ("key", "schedule_for")
  DO UPDATE SET
    data = excluded.data;

-- DeleteSiteData only removes future scheduled items.
-- To remove past scheduled items, use CleanSiteData
-- name: DeleteSiteData :exec
DELETE FROM site_data
WHERE "key" = @key
  AND "schedule_for" > (CURRENT_TIMESTAMP + '5 minutes'::interval);
