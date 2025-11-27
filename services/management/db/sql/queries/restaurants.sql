-- name: InsertRestaurant :one
INSERT INTO management.restaurants (id, name, address)
VALUES ($1, $2, $3)
RETURNING id, name, address, created_at, updated_at, deleted_at;

-- name: UpdateRestaurant :one
UPDATE management.restaurants
SET
    name = COALESCE(sqlc.narg(name), name),
    address = COALESCE(sqlc.narg(address), address),
    deleted_at = CASE
        WHEN sqlc.narg(delete_flag)::boolean IS NULL THEN deleted_at
        WHEN sqlc.narg(delete_flag) = TRUE THEN NOW()
        WHEN sqlc.narg(delete_flag) = FALSE THEN NULL
    END,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, address, created_at, updated_at, deleted_at;

-- name: InsertRestaurantManager :one
INSERT INTO management.restaurants_managers (id, user_id, restaurant_id)
VALUES ($1, $2, $3)
RETURNING id, user_id, restaurant_id, created_at, updated_at;

-- name: InsertRestaurantMenu :one
INSERT INTO management.menus (id, restaurant_id)
VALUES ($1, $2)
RETURNING id, restaurant_id, created_at, updated_at;

-- name: GetRestaurants :many
-- Get paginated list of restaurants
SELECT
    id,
    name,
    address,
    created_at
FROM management.restaurants
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: GetRestaurantByID :one
-- Get a single restaurant by its ID
SELECT
    id,
    name,
    address,
    created_at
FROM management.restaurants
WHERE id = $1;

-- name: IsUserRestaurantManager :one
-- Check if a user is a manager for a given restaurant
SELECT id, user_id, restaurant_id, created_at, updated_at
FROM management.restaurants_managers
WHERE user_id = $1
  AND restaurant_id = $2;

-- name: CreateTable :one
INSERT INTO management.tables (
    id,
    restaurant_id,
    name,
    capacity
) VALUES ($1, $2, $3, $4)
RETURNING id, restaurant_id, name, capacity;
