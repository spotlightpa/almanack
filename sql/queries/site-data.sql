-- name: GetSiteData :one
SELECT
  "data"
FROM
  site_data
WHERE
  "key" = $1;

-- name: SetSiteData :exec
UPDATE
  site_data
SET
  "data" = $2
WHERE
  "key" = $1;
