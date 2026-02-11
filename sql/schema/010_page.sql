CREATE TABLE "page" (
  id bigserial PRIMARY KEY,
  path text NOT NULL UNIQUE,
  frontmatter jsonb NOT NULL DEFAULT '{}'::jsonb,
  body text NOT NULL DEFAULT '',
  schedule_for timestamptz,
  last_published timestamptz,
  created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER row_updated_at_on_page_trigger_
  BEFORE UPDATE ON "page"
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE "page";
