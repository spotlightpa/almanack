-- name: UpdateArcArticles :exec
WITH arc_table AS (
    SELECT
        jsonb_array_elements(@arc_items::jsonb) AS article_data)
INSERT INTO article (arc_id, arc_data)
SELECT
    article_data ->> '_id',
    article_data
FROM
    arc_table
ON CONFLICT (arc_id)
    DO UPDATE SET
        arc_data = excluded.arc_data
    WHERE
        article.status <> 'A';

-- name: UpdateArcArticle :one
UPDATE
    article
SET
    arc_data = $2
WHERE
    arc_id = $1
RETURNING
    *;

-- name: UpdateAlmanackArticle :one
UPDATE
    article
SET
    status = $2,
    note = $3
WHERE
    arc_id = $1
RETURNING
    *;

-- name: UpdateSpotlightPAArticle :one
UPDATE
    article
SET
    spotlightpa_path = $2,
    spotlightpa_data = $3,
    schedule_for = $4,
    last_published = $5
WHERE
    arc_id = $1
RETURNING
    *;

-- name: PopScheduled :many
UPDATE
    article
SET
    last_published = CURRENT_TIMESTAMP
WHERE
    last_published IS NULL
    AND schedule_for < CURRENT_TIMESTAMP
RETURNING
    *;

-- name: GetArticle :one
SELECT
    *
FROM
    article
WHERE
    arc_id = $1;

-- name: ListUpcoming :many
SELECT
    *
FROM
    article
ORDER BY
    arc_data -> 'last_updated_date' DESC;

-- name: ListAllArticles :many
SELECT
    *
FROM
    article
ORDER BY
    arc_data -> 'last_updated_date' DESC;

-- name: ListAvailableArticles :many
SELECT
    *
FROM
    article
WHERE
    status <> 'U'
ORDER BY
    CASE status
    WHEN 'P' THEN
        '0'
    WHEN 'A' THEN
        '1'
    END ASC,
    arc_data -> 'last_updated_date' DESC;
