-- name: GetArcByArcID :one
SELECT
  *
FROM
  arc
WHERE
  arc_id = $1;
