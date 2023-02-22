ALTER TABLE "page"
  ADD COLUMN "publication_date" timestamptz GENERATED ALWAYS AS
    (iso_to_timestamptz (frontmatter ->> 'published')) STORED;

---- create above / drop below ----
ALTER TABLE "page"
  DROP COLUMN "publication_date";
