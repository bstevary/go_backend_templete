-- name: CreatePermission :exec
INSERT INTO permissions (
    name,
    code,
    description,
    resource,
    privacy
) VALUES ( $1, $2, $3, $4, $5
);


-- name: UpdatePermission :exec
UPDATE permissions
SET 
    updated = NOW(),
    name = COALESCE(sqlc.narg(name), name),
    code = COALESCE(sqlc.narg(code), code),
    privacy = COALESCE(sqlc.narg(privacy), privacy),
    resource = COALESCE(sqlc.narg(resource), resource),
    description = COALESCE(sqlc.narg(description), description)
WHERE
id = $1;

-- name: ListPermissions :many
SELECT 
  *, COUNT(*) OVER () AS count
FROM permissions 
WHERE privacy IN (sqlc.slice('privacy'))
ORDER BY id DESC
LIMIT $1 OFFSET $2;



-- name: GetPermission :many
SELECT *
FROM permissions
WHERE privacy IN (sqlc.slice('privacy'));


-- name: DeletePermission :exec
DELETE FROM permissions
WHERE id = $1;

