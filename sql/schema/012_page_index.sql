CREATE INDEX page_published ON page ((frontmatter ->> 'published'));

---- create above / drop below ----
DROP INDEX page_published;
