-- name: InsertMenuCategory :one
INSERT INTO management.categories (id, menu_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, menu_id, name, description, created_at, updated_at;
