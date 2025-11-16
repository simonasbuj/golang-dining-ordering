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
    token_version,
    name,
    lastname,
    role,
    is_active,
    created_at,
    updated_at,
    deleted_at
FROM auth.users
WHERE email = $1;

-- name: IncrementTokenVersion :one
UPDATE auth.users
SET token_version = token_version + 1
WHERE id = $1
RETURNING token_version;

-- name: GetTokenVersionByUserID :one
SELECT token_version
FROM auth.users
WHERE id = $1;