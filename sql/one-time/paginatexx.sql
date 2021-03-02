with
latest AS (
    SELECT
        to_timestamp(arc_data ->> 'last_updated_date'::text,
            -- ISO date
            'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS last_updated_date,
        id
    FROM
        article
    ORDER BY
        last_updated_date DESC,
        id DESC
    LIMIT 1
),
starting_point AS (
    SELECT
        to_timestamp(arc_data ->> 'last_updated_date'::text,
            -- ISO date
            'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS last_updated_date,
        id
    FROM
        article
    WHERE
        id = coalesce(@paginate_from_id
            ,
            (
                SELECT
                    id FROM latest)))
SELECT
    *
FROM
    article
WHERE (to_timestamp(arc_data ->> 'last_updated_date'::text,
            -- ISO date
            'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone, id) <= (
        SELECT
            *
        FROM
            starting_point)
ORDER BY
    to_timestamp(arc_data ->> 'last_updated_date'::text,
            -- ISO date
            'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone DESC,
    id DESC
LIMIT 50;


WITH latest AS (
  SELECT
    to_timestamp(arc_data ->> 'last_updated_date'::text,
      -- ISO date
      'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS last_updated_date,
    id
  FROM
    article
  ORDER BY
    last_updated_date DESC,
    id DESC
  LIMIT 1
),
starting_point AS (
  SELECT
    to_timestamp(arc_data ->> 'last_updated_date'::text,
      -- ISO date
      'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone AS last_updated_date,
    id
  FROM
    article
  WHERE
    article.id = coalesce(@paginate_from_id, (
        SELECT
          id FROM latest)))
SELECT
  *
FROM
  article
WHERE (to_timestamp(arc_data ->> 'last_updated_date'::text,
    -- ISO date
    'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone,
  id) <= (
    SELECT
      *
    FROM
      starting_point)
ORDER BY
  to_timestamp(arc_data ->> 'last_updated_date'::text,
    -- ISO date
    'YYYY-MM-DD"T"HH24:MI:SS"Z"')::timestamp WITH time zone DESC,
  id DESC
LIMIT 50;
