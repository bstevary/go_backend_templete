-- name: CreateUserRole :exec
INSERT INTO authorities (
    user_id,
    role,
    reference
) VALUES ($1, $2, $3);

-- name: GetUserRoles :many
SELECT r.name, r.reference
FROM authorities ur
JOIN roles r ON ur.role = r.id
WHERE ur.user_id = $1 AND r.reference = COALESCE(sqlc.narg(reference), r.reference);


-- name: DeleteUserRole :exec
DELETE FROM authorities
WHERE role = $1 AND user_id = $2 AND reference = COALESCE(sqlc.narg(reference), reference);


-- name: GetUserRolesByBranch :many
SELECT 
  COALESCE(ur.reference, 0) AS reference, 
  r.name AS role,
  p.code AS permission
FROM authorities ur
JOIN roles r ON ur.role = r.id
JOIN privileges rp ON rp.role = r.id
JOIN permissions p ON p.id = rp.permission
WHERE ur.user_id = $1
ORDER BY ur.reference, r.name;



-- name: AssingnNamedRoleToUser :exec
INSERT INTO authorities (user_id, role, reference)
SELECT $1, id, $2
FROM roles
WHERE name = $1 AND reference IS NULL;


-- -- name: GetUserProfileWithRoles :one
-- SELECT 
--   c.*,  u.id,  u.email,
--   u.mobile, u.username,
--   u.is_active, u.is_locked,
--   u.is_verified, u.mfa_enabled,
--   u.timestamp, u.updated,
--   u.failed_attempts,
--   u.last_password_change, 
--   mb.occupation, mb.device,
--   mb.worship,mb.latitude,
--   mb.longitude, mb.address,
--   mb.province, mb.district,
--   mb.city, mb.region,
--   mb.signature, mb.marital_status,
--   mb.id_front, mb.id_back,
--   mb.spouse_name, mb.spouse_number,
--   (
--     SELECT JSON_ARRAYAGG(
--       JSON_OBJECT(
--         'id', r.id,
--         'name', r.name,
--         'description', r.description,
--         'reference', r.reference,
--         'tag', r.tag,
--         'timestamp', r.timestamp,
--         'updated', r.updated
--       )
--     )
--     FROM roles r
--     JOIN authorities a ON a.role = r.id
--     WHERE a.user_id = u.id     
--   ) AS roles
-- FROM users u
-- JOIN members c ON u.id = c.id
-- LEFT JOIN member_bios mb ON c.id = mb.id
-- WHERE c.id = $1;


