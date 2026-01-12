-- name: GetRedirect :one
SELECT
  *
FROM
  "redirect"
WHERE
  "from" = $1;
