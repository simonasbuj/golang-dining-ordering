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

-- name: UpdateItem :one
UPDATE management.items
SET
    category_id    = $2,
    name           = $3,
    description    = $4,
    price_in_cents = $5,
    is_available   = $6,
    image_path     = CASE
                        WHEN sqlc.narg(image_path)::text IS NULL OR sqlc.narg(image_path) = '' THEN image_path
                        ELSE sqlc.narg(image_path)
                     END,
    updated_at     = NOW()
WHERE id = $1
RETURNING *;

-- name: GetItemByID :one
SELECT
    *
FROM management.items
WHERE id = $1;

-- name: GetMenuCategoriesWithItems :many
SELECT json_build_object(
    'categories', json_agg(
        json_build_object(
            'id', c.id,
            'name', c.name,
            'description', c.description,
            'created_at', c.created_at,
            'items', COALESCE(
                (SELECT json_agg(
                    json_build_object(
                        'id', i.id,
                        'category_id', i.category_id,
                        'name', i.name,
                        'description', i.description,
                        'price_in_cents', i.price_in_cents,
                        'image_path', i.image_path,
                        'is_available', i.is_available,
                        'created_at', i.created_at
                    )
                ) FROM management.items i WHERE i.category_id = c.id),
                '[]'::json
            )
        )
    )
) AS result
FROM management.categories c
WHERE c.menu_id = $1;

