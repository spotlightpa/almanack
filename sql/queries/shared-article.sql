-- name: UpdateSharedArticle :one
UPDATE
  shared_article
SET
  "embargo_until" = @embargo_until,
  "status" = @status,
  "note" = @note,
  "publication_date" = @publication_date,
  "internal_id" = @internal_id,
  "byline" = @byline,
  "budget" = @budget,
  "hed" = @hed,
  "description" = @description,
  "lede_image" = @lede_image,
  "lede_image_credit" = @lede_image_credit,
  "lede_image_description" = @lede_image_description,
  "lede_image_caption" = @lede_image_caption
WHERE
  id = @id
RETURNING
  *;

-- name: UpdateSharedArticlePage :one
UPDATE
  shared_article
SET
  page_id = @page_id
WHERE
  id = @shared_article_id
RETURNING
  *;

-- name: GetSharedArticleByID :one
SELECT
  *
FROM
  shared_article
WHERE
  id = $1;

-- name: GetSharedArticleBySource :one
SELECT
  *
FROM
  shared_article
WHERE
  source_type = $1
  AND source_id = $2;

-- name: ListSharedArticles :many
SELECT
  *
FROM
  shared_article
ORDER BY
  publication_date DESC
LIMIT $1 OFFSET $2;

-- name: ListSharedArticlesWhereActive :many
SELECT
  *
FROM
  shared_article
WHERE
  status <> 'U'
ORDER BY
  publication_date DESC
LIMIT $1 OFFSET $2;

-- name: UpsertSharedArticleFromArc :one
INSERT INTO shared_article (status, source_type, source_id, raw_data,
  publication_date, budget, description, hed, internal_id)
SELECT
  'U',
  'arc',
  arc.arc_id,
  arc.raw_data,
  iso_to_timestamptz ( --
    arc.raw_data -> 'planning' -> 'scheduling' ->> 'planned_publish_date'),
  arc.raw_data -> 'planning' ->> 'budget_line',
  arc.raw_data -> 'description' ->> 'basic',
  arc.raw_data -> 'headlines' ->> 'basic',
  arc.raw_data ->> 'slug'
FROM
  arc
WHERE
  arc_id = $1
ON CONFLICT (source_type,
  source_id)
  DO UPDATE SET
    raw_data = excluded.raw_data,
    "publication_date" = iso_to_timestamptz ( --
      excluded.raw_data -> 'planning' -> 'scheduling' ->> 'planned_publish_date'),
    "budget" = excluded.raw_data -> 'planning' ->> 'budget_line',
    "description" = excluded.raw_data -> 'description' ->> 'basic',
    "hed" = excluded.raw_data -> 'headlines' ->> 'basic',
    "internal_id" = excluded.raw_data ->> 'slug'
  RETURNING
    *;

-- name: UpsertSharedArticleFromGDocs :one
INSERT INTO shared_article (status, source_type, source_id, raw_data,
  internal_id, byline, budget, hed, description, lede_image, lede_image_credit,
  lede_image_description, lede_image_caption)
  VALUES ('U', 'gdocs', @external_id, @raw_data::jsonb,
    @internal_id, @byline, @budget, @hed, @description, @lede_image,
    @lede_image_credit, @lede_image_description, @lede_image_caption)
ON CONFLICT (source_type, source_id)
  DO UPDATE SET
    raw_data = excluded.raw_data
  RETURNING
    *;

-- name: UpdateSharedArticleFromGDocs :one
UPDATE
  shared_article
SET
  "raw_data" = @raw_data,
  "internal_id" = @internal_id,
  "byline" = @byline,
  "budget" = @budget,
  "hed" = @hed,
  "description" = @description,
  "lede_image" = @lede_image,
  "lede_image_credit" = @lede_image_credit,
  "lede_image_description" = @lede_image_description,
  "lede_image_caption" = @lede_image_caption
WHERE
  source_type = 'gdocs'
  AND source_id = @external_id
RETURNING
  *;
