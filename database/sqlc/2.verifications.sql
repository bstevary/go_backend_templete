-- name: CreateVerification :exec
INSERT INTO verifications (
    expiry,
    user_id,
    otp,
    channel,
    purpose
) VALUES (
    NOW() + INTERVAL '10 minutes', $1, $2, $3, $4
);

-- name: UpdateVerification :exec
UPDATE verifications
SET
    is_used = TRUE
WHERE otp = $1
    AND is_used = FALSE
    AND expiry > NOW()
;

-- name: GetVerificationUser :one
 SELECT user_id FROM verifications 
 WHERE otp = $1
    AND is_used = FALSE
    AND expiry > NOW()
 LIMIT 1;