-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id,
    user_id,
    token_hash,
    expires_at,
    created_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    id,
    user_id,
    token_hash,
    expires_at,
    created_at,
    revoked_at;

-- name: GetRefreshTokenByHash :one
SELECT
    id,
    user_id,
    token_hash,
    expires_at,
    created_at,
    revoked_at
FROM refresh_tokens
WHERE token_hash = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $2
WHERE id = $1
  AND revoked_at IS NULL;
