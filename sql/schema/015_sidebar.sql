WITH edpicks AS (
  SELECT
    data
  FROM
    site_data
  WHERE
    key = 'data/editorsPicks.json'
  ORDER BY
    published_at DESC
  LIMIT 1
),
items AS (
  SELECT
    jsonb_array_elements_text(data -> 'sidebarPicks') AS item
FROM
  edpicks
),
blob AS (
  SELECT
    json_build_object('items', (
        SELECT
          json_agg(row_to_json("items")::jsonb || '
            {
              "title": "Editor’s Pick",
              "labelColor": "#ff6c36",
              "linkColor": "#000000",
              "backgroundColor": "#f5f5f5"
            }
    '::jsonb)
        FROM "items")) AS data)
  INSERT INTO site_data (key, data, schedule_for, published_at)
  SELECT
    'data/sidebar.json',
    data,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  FROM
    blob
  RETURNING
    *;

---- create above / drop below ----
DELETE FROM site_data
WHERE key = 'data/sidebar.json';
