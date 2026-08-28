-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash
)
VALUES (
    $1,
    $2
)
RETURNING
    id,
    email,
    password_hash,
    email_verified,
    created_at,
    updated_at;


-- name: GetUserByID :one
SELECT
    id,
    email,
    password_hash,
    email_verified,
    created_at,
    updated_at
FROM users
WHERE id = $1;


-- name: GetUserByEmail :one
SELECT
    id,
    email,
    password_hash,
    email_verified,
    created_at,
    updated_at
FROM users
WHERE lower(email) = lower($1)
LIMIT 1;


-- name: ActivateUser :exec
UPDATE users
SET
    email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1
  AND email_verified = FALSE;


-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;