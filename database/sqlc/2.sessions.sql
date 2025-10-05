-- name: CreateSession :exec
INSERT INTO
  sessions (
    id,
    user_id,
    refresh_token,
    user_agent,
    client_ip,
    platform,
    device,
    expiry
  )
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetSession :one
SELECT  *
FROM  sessions
WHERE  id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: UpdateSession :exec
UPDATE sessions
SET
  refresh_token = COALESCE(sqlc.narg (refresh_token)),
  expiry = COALESCE(sqlc.narg (expiry)),
  is_blocked = COALESCE(sqlc.narg (is_blocked))
WHERE
  id = $1;