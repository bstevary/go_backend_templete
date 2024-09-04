-- name: CreateSession :one
INSERT INTO sessions (
  id,
  email,
  refresh_token,
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: UpdateSession :exec
UPDATE sessions
SET
  refresh_token = COALESCE(sqlc.narg( refresh_token)),
  expires_at = COALESCE(sqlc.narg( expires_at)),
  is_blocked = COALESCE(sqlc.narg( is_blocked))
WHERE email = sqlc.narg(email) 
 RETURNING *;


