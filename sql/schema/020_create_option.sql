CREATE TABLE "option" (
  "id" bigserial PRIMARY KEY,
  "key" text NOT NULL,
  "value" text NOT NULL DEFAULT ''
);

---- create above / drop below ----
DROP TABLE "option";
