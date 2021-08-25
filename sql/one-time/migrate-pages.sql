INSERT INTO page (file_path, last_published)
SELECT
  spotlightpa_path,
  CURRENT_TIMESTAMP
FROM
  article
WHERE
  article.spotlightpa_path IS NOT NULL
  AND article.last_published IS NOT NULL
RETURNING
  *;
