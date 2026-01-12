CREATE TABLE redirect (
  "id" bigserial PRIMARY KEY,
  "from" text NOT NULL UNIQUE,
  "to" text NOT NULL DEFAULT '',
  "roles" text[] NOT NULL DEFAULT '{}',
  "code" integer NOT NULL,
  "created_at" timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

---- create above / drop below ----
DROP TABLE redirect;
