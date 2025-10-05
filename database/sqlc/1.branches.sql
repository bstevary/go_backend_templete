-- name: CreateBranch :exec
INSERT INTO branches (
    name,
    code,
    description,
    country,
    region,
    province,
    district,
    city,
    address,
    contact,
    email,
    other,
    is_active,
    subscription
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);


-- name: GetBranchByID :one
SELECT * FROM branches WHERE id = $1;

-- name: UpdateBranch :exec
UPDATE branches
SET 
    updated_at = NOW(),
    name = COALESCE(sqlc.narg(name), name),
    code = COALESCE(sqlc.narg(code), code),
    description = COALESCE(sqlc.narg(description), description),
    country = COALESCE(sqlc.narg(country), country),
    region = COALESCE(sqlc.narg(region), region),
    province = COALESCE(sqlc.narg(province), province),
    district = COALESCE(sqlc.narg(district), district),
    city = COALESCE(sqlc.narg(city), city),
    address = COALESCE(sqlc.narg(address), address),
    contact = COALESCE(sqlc.narg(contact), contact),
    email = COALESCE(sqlc.narg(email), email),
    other = COALESCE(sqlc.narg(other), other),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    subscription = COALESCE(sqlc.narg(subscription), subscription)    
WHERE id = $1;


-- name: DeleteBranch :exec
DELETE FROM branches WHERE id = $1;

-- name: ListBranches :many
SELECT * FROM branches 
ORDER BY id DESC
LIMIT $1 OFFSET $2;