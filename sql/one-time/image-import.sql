WITH input AS (
    SELECT
        jsonb_array_elements(
            -- insert JSON here
            $$ REPLACEME $$
            --
::jsonb) AS data)
INSERT INTO image (path, credit, description, created_at)
SELECT
    data ->> 'image' AS path,
    min(data ->> 'image-credit'),
    coalesce(min(data ->> 'image-description'), ''),
    min(to_timestamp(data ->> 'created-at', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
FROM
    input
GROUP BY
    path
ON CONFLICT (path)
    DO NOTHING;

