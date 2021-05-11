INSERT INTO image_type (name, mime, extensions)
  VALUES
    --
    ('tiff', 'image/tiff', '{tif,tiff}');

---- create above / drop below ----
DELETE FROM image_type
WHERE name = 'tiff';
