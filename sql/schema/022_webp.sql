INSERT INTO image_type (name, mime, extensions)
  VALUES ('webp', 'image/webp', '{webp}');

---- create above / drop below ----

DELETE FROM image_type WHERE name = 'webp';
