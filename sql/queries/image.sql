-- name: ListImages :many
SELECT
    *
FROM
    image
ORDER BY
    created_at DESC
LIMIT 100;

-- name: InsertImage :one
INSERT INTO image (path, credit, description, src_url)
    VALUES (@path, @credit, @description, @src_url)
ON CONFLICT (path)
    DO UPDATE SET
        credit = CASE WHEN credit = '' THEN
            excluded.credit
        ELSE
            credit
        END, description = CASE WHEN description = '' THEN
            excluded.description
        ELSE
            description
        END, src_url = CASE WHEN src_url = '' THEN
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

