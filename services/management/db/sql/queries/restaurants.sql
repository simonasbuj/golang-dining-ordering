-- name: InsertRestaurant :one
INSERT INTO management.restaurants (id, name, address)
VALUES ($1, $2, $3)
RETURNING id, name, address, created_at, updated_at, deleted_at;

-- name: InsertRestaurantManager :one
INSERT INTO management.restaurant_managers (id, user_id, restaurant_id)
VALUES ($1, $2, $3)
RETURNING id, user_id, restaurant_id, created_at, updated_at;
