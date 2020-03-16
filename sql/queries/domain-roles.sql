-- name: GetRolesForDomain :one
SELECT
    roles
FROM
    domain_roles
WHERE
    domain ILIKE $1;

-- name: SetRolesForDomain :one
INSERT INTO domain_roles (domain, roles)
    VALUES ($1, $2)
ON CONFLICT (lower(domain))
    DO UPDATE SET
        roles = $2
    RETURNING
        *;
