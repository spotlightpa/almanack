-- name: CreateGDocsDoc :one
INSERT INTO g_docs_doc ("g_docs_id", "document")
  VALUES (@g_docs_id, @document)
RETURNING
  *;

-- name: GetGDocsByID :one
SELECT
  *
FROM
  g_docs_doc
WHERE
  id = $1;

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
  "embeds" = @embeds,
  "rich_text" = @rich_text,
  "raw_html" = @raw_html,
  "article_markdown" = @article_markdown,
  "word_count" = @word_count,
  "processed_at" = CURRENT_TIMESTAMP
WHERE
  id = @id
RETURNING
  *;

-- name: UpsertGDocsImage :exec
INSERT INTO g_docs_image (g_docs_id, doc_object_id, image_id)
  VALUES (@g_docs_id, @doc_object_id, @image_id)
ON CONFLICT (g_docs_id, doc_object_id)
  DO NOTHING;

-- name: ListGDocsImagesByGDocsID :many
SELECT
  "doc_object_id",
  "path"::text,
  "type"::text
FROM
  g_docs_image
  LEFT JOIN image ON (image_id = image.id)
WHERE
  g_docs_id = $1;
