-- name: CreateActivateAccountEmail :one
INSERT INTO "varify_email" (
    user_id,
    email,
    secret_code
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateActivateAccountEmail :one
UPDATE "varify_email"
SET
    is_used = TRUE
WHERE secret_code = @secret_code
    AND is_used = FALSE
    AND expired_at > now()
RETURNING *;
