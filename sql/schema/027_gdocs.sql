CREATE TABLE g_docs_doc (
  "id" bigserial PRIMARY KEY,
  "g_docs_id" text NOT NULL,
  "document" jsonb NOT NULL DEFAULT '{}' ::jsonb,
  "embeds" jsonb DEFAULT '[]' ::jsonb,
  "rich_text" text NOT NULL DEFAULT '',
  "raw_html" text NOT NULL DEFAULT '',
  "article_markdown" text NOT NULL DEFAULT '',
  "word_count" integer NOT NULL DEFAULT 0,
  "processed_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX "g_docs_doc_processed_at" ON "g_docs_doc" ("processed_at");

CREATE TABLE g_docs_image (
  "id" bigserial PRIMARY KEY,
  "g_docs_id" text NOT NULL,
  "doc_object_id" text NOT NULL,
  "src_url" text NOT NULL,
  "image_id" bigint REFERENCES image (id),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX "gdocs_image_unique_gdocs_id_gdoc_object_id" ON
  "g_docs_image" ("g_docs_id", "doc_object_id");

CREATE TRIGGER row_updated_at_on_g_docs_trigger_
  BEFORE UPDATE ON g_docs_image
  FOR EACH ROW
  EXECUTE PROCEDURE update_row_updated_at_function_ ();

---- create above / drop below ----
DROP TABLE g_docs_doc;

DROP TABLE g_docs_image;
