-- name: InsertMenuCategory :one
INSERT INTO management.categories (id, menu_id, name, description)
VALUES ($1, $2, $3, $4)
RETURNING id, menu_id, name, description, created_at, updated_at;

-- name: InsertMenuItem :one
INSERT INTO management.items (
    id,
    category_id,
    name,
    description,
    price_in_cents,
    is_available,
    image_path
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;
