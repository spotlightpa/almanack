-- name: CreateGDocsDoc :one
INSERT INTO g_docs_doc ("external_id", "document")
  VALUES (@external_id, @document)
RETURNING
  *;

-- name: GetGDocsByID :one
SELECT
  *
FROM
  g_docs_doc
WHERE
  id = $1;

-- name: GetGDocsByExternalIDWhereProcessed :one
SELECT
  *
FROM
  g_docs_doc
WHERE
  external_id = $1
  AND processed_at IS NOT NULL
ORDER BY
  processed_at DESC
LIMIT 1;

-- name: ListGDocsWhereUnprocessed :many
SELECT
  *
FROM
  g_docs_doc
WHERE
  processed_at IS NULL;

-- name: UpdateGDocsDoc :one
UPDATE
  g_docs_doc
SET
  "metadata" = @metadata,
  "embeds" = @embeds,
  "rich_text" = @rich_text,
  "raw_html" = @raw_html,
  "article_markdown" = @article_markdown,
  "word_count" = @word_count,
  "warnings" = @warnings,
  "processed_at" = CURRENT_TIMESTAMP
WHERE
  id = @id
RETURNING
  *;

-- name: UpsertGDocsImage :exec
INSERT INTO g_docs_image (external_id, doc_object_id, image_id)
  VALUES (@external_id, @doc_object_id, @image_id)
ON CONFLICT (external_id, doc_object_id)
  DO NOTHING;

-- name: ListGDocsImagesByExternalID :many
SELECT
  "doc_object_id",
  "path"::text,
  "type"::text
FROM
  g_docs_image
  LEFT JOIN image ON (image_id = image.id)
WHERE
  external_id = $1;

-- name: DeleteGDocsDocWhereUnunused :exec
DELETE FROM g_docs_doc
WHERE id NOT IN (
    SELECT
      raw_data::bigint
    FROM
      shared_article
    WHERE
      source_type = 'gdocs')
  AND processed_at < CURRENT_TIMESTAMP - interval '1 hour';
