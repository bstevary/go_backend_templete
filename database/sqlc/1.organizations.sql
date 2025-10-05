-- name: CreateOrganization :exec
INSERT INTO organizations (
    branch,
    name,
    code,
    description,
    is_active,
    contact,
    email,
    country,
    region,
    province,
    district,
    city,
    address,
    other
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: UpdateOrganization :exec
UPDATE organizations
SET 
    updated_at = NOW(),
    branch = COALESCE(sqlc.narg(branch), branch),
    name = COALESCE(sqlc.narg(name), name),
    code = COALESCE(sqlc.narg(code), code),
    description = COALESCE(sqlc.narg(description), description),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    contact = COALESCE(sqlc.narg(contact), contact),
    email = COALESCE(sqlc.narg(email), email),
    country = COALESCE(sqlc.narg(country), country),
    region = COALESCE(sqlc.narg(region), region),
    province = COALESCE(sqlc.narg(province), province),
    district = COALESCE(sqlc.narg(district), district),
    city = COALESCE(sqlc.narg(city), city),
    address = COALESCE(sqlc.narg(address), address),
    other = COALESCE(sqlc.narg(other), other)

WHERE id = $1;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations 
ORDER BY id DESC
LIMIT $1 OFFSET $2;