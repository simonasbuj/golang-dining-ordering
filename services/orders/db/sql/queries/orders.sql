-- name: GetCurrentOrder :one
SELECT id
FROM orders.orders
WHERE table_id = $1;
