-- name: CreatePromotion :one
INSERT INTO "promotion" ("name", "description", "link", "width", "height", "image_urls")
  VALUES (@name, @description, @link, @width, @height, @image_urls)
RETURNING
  *;

-- name: UpdatePromotion :one
UPDATE
  "promotion"
SET
  "name" = @name,
  "description" = @description,
  "link" = @link,
  "width" = @width,
  "height" = @height,
  "image_urls" = @image_urls,
  "updated_at" = CURRENT_TIMESTAMP
WHERE
  "id" = @id
RETURNING
  *;

-- name: ListPromotionByUpdated :many
SELECT
  *
FROM
  "promotion"
WHERE ("promotion"."width" = @width
  OR "promotion"."width" = 0
  OR @width = 0)
AND ("promotion"."height" = @height
  OR "promotion"."height" = 0
  OR @height = 0)
ORDER BY
  "updated_at" DESC
LIMIT $1 OFFSET $2;

-- name: ListPromotionByFTS :many
WITH tsq AS (
  SELECT
    websearch_to_tsquery('english', @text::text) AS q
)
SELECT
  "promotion".*
FROM
  "promotion"
  INNER JOIN tsq ON fts_doc_en @@ tsq.q
WHERE ("promotion"."width" = @width
  OR "promotion"."width" = 0
  OR @width = 0)
AND ("promotion"."height" = @height
  OR "promotion"."height" = 0
  OR @height = 0)
ORDER BY
  ts_rank(fts_doc_en, tsq.q) DESC
LIMIT $1 OFFSET $2;

-- name: DeletePromotion :execrows
DELETE FROM "promotion"
WHERE "id" = @id;
