CREATE OR REPLACE FUNCTION iso_to_timestamptz (text)
  RETURNS timestamptz
  AS $$
  SELECT
    to_timestamp($1, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamptz
$$
LANGUAGE sql
IMMUTABLE;

---- create above / drop below ----
DROP FUNCTION IF EXISTS iso_to_timestamptz (text);
