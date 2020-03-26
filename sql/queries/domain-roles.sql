-- name: GetRolesForDomain :one
SELECT
    roles
FROM
    domain_roles
WHERE
    "domain" ILIKE $1;

-- name: SetRolesForDomain :one
INSERT INTO domain_roles ("domain", roles)
    VALUES ($1, $2)
ON CONFLICT (lower("domain"))
    DO UPDATE SET
        roles = $2
    RETURNING
        *;

-- name: AppendRoleToDomain :one
INSERT INTO domain_roles ("domain", roles)
    VALUES (@domain, ARRAY[@role::text])
ON CONFLICT (lower("domain"))
    DO UPDATE SET
        roles = CASE WHEN NOT (domain_roles.roles::text[] @> ARRAY[@role]) THEN
            domain_roles.roles::text[] || ARRAY[@role]
        ELSE
            domain_roles.roles
        END
    RETURNING
        *;

-- name: ListDomainsWithRole :many
SELECT
    "domain"
FROM
    "domain_roles"
WHERE
    "roles" @> ARRAY[@role::text]
ORDER BY
    "domain" ASC;

