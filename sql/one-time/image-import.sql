WITH input AS (
  SELECT
    jsonb_array_elements(
      -- insert JSON here
      $$ REPLACEME $$
      --
::jsonb) AS data)
INSERT INTO image (path, credit, description, is_uploaded, created_at)
SELECT
  data ->> 'image' AS path,
  min(data ->> 'image-credit'),
  coalesce(min(data ->> 'image-description'), ''),
  TRUE,
  min(to_timestamp(data ->> 'created-at', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
FROM
  input
GROUP BY
  path
ON CONFLICT (path)
  DO UPDATE SET
    description = coalesce(image.description, excluded.description),
    created_at = least (image.created_at, excluded.created_at),
    is_uploaded = TRUE
  RETURNING
    *;
