-- name: CreateUser :one
INSERT INTO users (
    first_name,
    last_name,
    email,
    gender, 
    hashed_password,
    date_of_birth
      )VALUES ($1, $2, $3, $4, $5, $6 
) RETURNING user_id, first_name,  last_name,  email, gender, date_of_birth;

-- name: UpdateUser :one
UPDATE users
SET 
  updated_at = NOW(),
  first_name = COALESCE(sqlc.narg(first_name), first_name),
  last_name = COALESCE(sqlc.narg(last_name), last_name),
  date_of_birth = COALESCE(sqlc.narg(date_of_birth), date_of_birth),  
  gender = COALESCE(
    sqlc.narg(gender),
    gender
  )
  WHERE email = $1
 RETURNING user_id, first_name,  last_name,  email, gender, date_of_birth;

-- name: AlterUserAccountStatus :one
UPDATE users
SET
  updated_at = NOW(),
  is_account_active = COALESCE(sqlc.narg(is_account_active), is_account_active),
  is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified)
WHERE user_id = $1
RETURNING *;

-- name: ListUsers :many
SELECT user_id, first_name,  last_name,  email, gender, date_of_birth,  is_email_verified, is_account_active, created_at  FROM users
WHERE  user_id > $1
ORDER BY user_id 
LIMIT $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: SelectUserByEmail :one
SELECT * FROM users
WHERE email = $1 
LIMIT 1;

-- name: DeleteUser :one
DELETE FROM users
WHERE email = $1
RETURNING user_id;








