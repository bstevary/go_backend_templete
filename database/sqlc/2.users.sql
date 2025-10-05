-- name: CreateUser :exec
INSERT INTO users (
    id,
    first_name,
    middle_name,
    last_name,
    other_name,
    email,
    contact,
    password,
    role,
    type
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateUserPersonalInfo :exec
UPDATE users
SET 
  updated_at = NOW(),
  first_name = COALESCE(sqlc.narg(first_name), first_name),
  middle_name = COALESCE(sqlc.narg(middle_name), middle_name),
  last_name = COALESCE(sqlc.narg(last_name), last_name),
  other_name = COALESCE(sqlc.narg(other_name), other_name)
WHERE id = $1;

-- name: UpdateUserContactInfo :exec
UPDATE users
SET 
  updated_at = NOW(),
  email = COALESCE(sqlc.narg(email), email),
  contact = COALESCE(sqlc.narg(contact), contact),
  is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified),
  is_contact_verified = COALESCE(sqlc.narg(is_contact_verified), is_contact_verified)
WHERE id = $1;

-- name: UpdateUserSecurityInfo :exec
UPDATE users
SET 
  updated_at = NOW(),
  password = COALESCE(sqlc.narg(password), password),
  is_locked = COALESCE(sqlc.narg(is_locked), is_locked),
  mfa_enabled = COALESCE(sqlc.narg(mfa_enabled), mfa_enabled),
  failed_attempts = COALESCE(sqlc.narg(failed_attempts), failed_attempts),
  last_password_change = COALESCE(sqlc.narg(last_password_change), last_password_change),
  last_security_check = COALESCE(sqlc.narg(last_security_check), last_security_check)
WHERE id = $1;

-- name: UpdateUserStatus :exec
UPDATE users
SET 
  updated_at = NOW(),
  is_active = COALESCE(sqlc.narg(is_active), is_active),
  accepted_terms = COALESCE(sqlc.narg(accepted_terms), accepted_terms)
WHERE id = $1;

-- name: UpdateUserRoleAndType :exec
UPDATE users
SET 
  updated_at = NOW(),
  role = COALESCE(sqlc.narg(role), role),
  type = COALESCE(sqlc.narg(type), type)
WHERE id = $1;

-- name: ListUsers :many
SELECT 
  id, first_name, middle_name, last_name, other_name, email, contact, 
  role, type, is_active, is_locked, mfa_enabled, failed_attempts, 
  is_email_verified, is_contact_verified, created_at, updated_at
FROM users
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: SelectUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: SelectUserByContact :one
SELECT * FROM users
WHERE contact = $1
LIMIT 1;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING id;








