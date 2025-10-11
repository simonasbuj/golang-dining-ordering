-- name: CreateUser :one
INSERT INTO users (
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
