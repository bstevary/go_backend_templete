-- name: CreateRole :exec
INSERT INTO roles (
  name,
  reference,
  scope,
  tag,
  description    
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: GetRoles :many
SELECT id, name, tag, reference, scope
FROM roles
WHERE (sqlc.narg(reference) IS NULL OR reference = sqlc.narg(reference));

-- name: UpdateRole :exec
UPDATE roles
SET 
    updated = NOW(),
    name = COALESCE(sqlc.narg(name), name),
    tag = COALESCE(sqlc.narg(tag), tag),
    reference = COALESCE(sqlc.narg(reference), reference),
    description = COALESCE(sqlc.narg(description), description)
WHERE id = $1;

-- name: ListAllRoles :many
SELECT *, CAST(COUNT(*) OVER () AS SIGNED) AS total_count
FROM roles
WHERE reference IS NULL 
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: ListRoles :many
SELECT *, CAST(COUNT(*) OVER () AS SIGNED) AS total_count
FROM roles
WHERE reference = $1
ORDER BY id DESC
LIMIT $2 OFFSET $3;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1
  AND (reference = $2);
