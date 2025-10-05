-- name: CreateRolePermission :exec
INSERT INTO privileges (
    role,
    permission    
) VALUES ($1, $2);

-- name: GetRolePermission :many
SELECT permission FROM privileges
WHERE  role = $1;


-- name: DeleteRolePermission :exec
DELETE FROM privileges
WHERE permission = $1 AND role = $2;


-- name: GetRole :one
SELECT  r.*
FROM roles r
WHERE (r.reference = $1 OR ($1 IS NULL AND r.reference IS NULL))
  AND r.id = $2;

-- name: GetRolePermissions :many
SELECT p.*
FROM permissions p
JOIN privileges rp ON rp.permission = p.id
WHERE rp.role = $1;


