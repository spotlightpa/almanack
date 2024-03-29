-- name: UpdateArc :exec
WITH arc_temp AS (
  SELECT
    jsonb_array_elements(@arc_items::jsonb) AS temp_data)
INSERT INTO arc (arc_id, raw_data)
SELECT
  temp_data ->> '_id',
  temp_data
FROM
  arc_temp
ON CONFLICT (arc_id)
  DO UPDATE SET
    raw_data = excluded.raw_data;

-- name: ListArcByLastUpdated :many
SELECT
  arc.*,
  shared_article.id AS shared_article_id,
  coalesce(shared_article.status, ''),
  shared_article.embargo_until
FROM
  arc
  LEFT JOIN shared_article ON (arc.arc_id = shared_article.source_id
      AND shared_article.source_type = 'arc')
  ORDER BY
    last_updated DESC
  LIMIT $1 OFFSET $2;

-- name: GetArcByArcID :one
SELECT
  *
FROM
  arc
WHERE
  arc_id = $1;
