-- name: GetOption :one
SELECT
  "value"
FROM
  "option"
WHERE
  key = $1;
