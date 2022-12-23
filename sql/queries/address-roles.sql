-- name: GetRolesForAddress :one
SELECT
  roles
FROM
  address_roles
WHERE
  "email_address" ILIKE $1;

-- name: UpsertRolesForAddress :one
INSERT INTO address_roles ("email_address", roles)
  VALUES ($1, $2)
ON CONFLICT (lower("email_address"))
  DO UPDATE SET
    roles = $2
  RETURNING
    *;

-- name: ListAddressesWithRole :many
SELECT
  "email_address"
FROM
  "address_roles"
WHERE
  "roles" @> ARRAY[@role::text]
ORDER BY
  "email_address" ASC;
