-- name: ListGDocsWhereUnprocessed :many
SELECT
  *
FROM
  g_docs_doc
WHERE
  processed_at IS NULL;
