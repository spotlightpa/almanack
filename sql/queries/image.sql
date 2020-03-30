-- name: ListImages :many
SELECT
  *
FROM
  image
ORDER BY
  created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateImage :one
INSERT INTO image ("path", "credit", "description", "src_url", "type")
  VALUES (@path, @credit, @description, @src_url, @type)
ON CONFLICT (path)
  DO UPDATE SET
    credit = excluded.credit, --
    description = excluded.description, --
    src_url = CASE WHEN src_url = '' THEN
      excluded.src_url
    ELSE
      src_url
    END
  RETURNING
    *;

-- name: GetImage :one
SELECT
  *
FROM
  image
WHERE
  path = $1;
