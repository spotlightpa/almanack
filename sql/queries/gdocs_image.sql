-- name: UpsertGDocsIDObjectID :exec
WITH objects_with_url (
  object_id,
  src_url
) AS (
  SELECT
    data ->> 0,
    data ->> 1
  FROM
    jsonb_array_elements(@object_url_pairs::jsonb) tjson (data))
  INSERT INTO g_docs_image (g_docs_id, doc_object_id, src_url)
  SELECT
    @g_docs_id,
    object_id,
    src_url
  FROM
    objects_with_url
  ON CONFLICT (g_docs_id,
    doc_object_id)
    DO UPDATE SET
      src_url = excluded.src_url;

-- name: ListGDocsImagesWhereUnset :many
SELECT
  *
FROM
  g_docs_image
WHERE
  image_id IS NULL;

-- name: UpdateGDocsImage :one
UPDATE
  g_docs_image
SET
  image_id = $1
WHERE
  id = $2
RETURNING
  *;
