-- name: GetCurrentOrder :one
SELECT id
FROM orders.orders
WHERE 
    table_id = $1 
    and status in ('open', 'locked')
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateOrder :one
INSERT INTO orders.orders (
    id,
    table_id,
    currency
) VALUES ($1, $2, $3)
RETURNING id;

-- name: GetTableCurrency :one
SELECT r.currency
FROM management.restaurants r
WHERE r.id = (select restaurant_id from management.tables t where t.id = $1);
