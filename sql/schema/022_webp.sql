INSERT INTO image_type (name, mime, extensions)
  VALUES
    --
    ('webp', 'image/webp', '{webp}'), --
    ('avif', 'image/avif', '{avif,avifs}'), --
    ('heic', 'image/heic', '{heic,heif}');

---- create above / drop below ----
DELETE FROM image_type
WHERE name IN ('webp', 'heic', 'avif');
