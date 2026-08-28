-- name: CreateEmailVerification :one
INSERT INTO email_verifications (
    user_id,
    code_hash,
    expires_at,
    attempts
)
VALUES (
    $1,
    $2,
    $3,
    0
)
RETURNING
    id,
    user_id,
    code_hash,
    expires_at,
    attempts,
    created_at,
    verified_at;


-- name: GetEmailVerificationByID :one
SELECT
    id,
    user_id,
    code_hash,
    expires_at,
    attempts,
    created_at,
    verified_at
FROM email_verifications
WHERE id = $1;


-- name: UpdateEmailVerificationCode :exec
UPDATE email_verifications
SET
    code_hash = $2,
    expires_at = $3,
    created_at = $4,
    attempts = 0,
    verified_at = NULL
WHERE id = $1;


-- name: IncrementEmailVerificationAttempts :exec
UPDATE email_verifications
SET attempts = attempts + 1
WHERE id = $1
  AND attempts < $2;


-- name: MarkEmailVerificationVerified :exec
UPDATE email_verifications
SET verified_at = $2
WHERE id = $1;
