CREATE TABLE youtube (
  "id" bigserial PRIMARY KEY,
  "external_id" text NOT NULL UNIQUE,
  "title" text NOT NULL DEFAULT '',
  "url" text NOT NULL DEFAULT '',
  "thumbnail_url" text NOT NULL DEFAULT '',
  "external_published_at" timestamp with time zone NOT NULL,
  "external_updated_at" timestamp with time zone NOT NULL,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

---- create above / drop below ----
DROP TABLE youtube;
