-- name: CreateUser :one
INSERT INTO auth.users (
    id,
    email,
    password_hash,
    name,
    lastname,
    role
)
VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING
    id,
    email,
    password_hash,
    name,
    lastname,
    role,
    created_at,
    updated_at,
    deleted_at;

-- name: GetUserByEmail :one
SELECT
    id,
    email,
    password_hash,
    name,
    lastname,
    role,
    is_active,
    created_at,
    updated_at,
    deleted_at
FROM auth.users
WHERE email = $1;

-- name: SaveRefreshToken :one
INSERT INTO auth.tokens (
    id,
    user_id,
    expires_at
) VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteRefreshToken :exec
DELETE FROM auth.tokens
WHERE user_id = $1
  AND id = $2;

-- name: GetRefreshToken :one
SELECT
    id,
    user_id,
    created_at
FROM auth.tokens
WHERE user_id = $1
  AND id = $2;
