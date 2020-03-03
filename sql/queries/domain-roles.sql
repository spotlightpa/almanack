-- name: GetRolesForDomain :one
SELECT
    roles
FROM
    domain_roles
WHERE
    DOMAIN LIKE $1;
