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

-- name: AddOrderItem :one
INSERT INTO orders.orders_items (
    id,
    order_id,
    item_id,
    item_name,
    price_in_cents
) VALUES ($1, $2, $3, $4, $5)
RETURNING order_id;

-- name: GetOrderItems :many
SELECT
    o.id,
    r.id as restaurant_id,
    o.status,
    o.currency,
    o.tip_amount_in_cents,
    i.id as order_item_id,
    i.item_id,
    i.item_name,
    i.price_in_cents
FROM orders.orders o
    LEFT JOIN orders.orders_items i ON o.id = i.order_id
    LEFT JOIN management.tables t on t.id = o.table_id
    LEFT JOIN management.restaurants r on r.id = t.restaurant_id
WHERE o.id = $1;

-- name: GetMenuItem :one
SELECT 
    i.id,
    m.id as restaurant_id,
    i.name,
    i.price_in_cents
FROM management.items i 
    LEFT JOIN management.categories c on c.id = i.category_id
    LEFT JOIN management.menus m on m.id = c.menu_id
WHERE i.id = $1;

-- name: DeleteOrderItem :exec
DELETE FROM orders.orders_items WHERE id = $1 and order_id = $2;

-- name: UpdateOrder :exec
UPDATE orders.orders
SET
    status = COALESCE(sqlc.narg(status), status),
    tip_amount_in_cents = COALESCE(sqlc.narg(tip_amount_in_cents), tip_amount_in_cents),
    updated_at = NOW()
WHERE id = $1;

-- name: IsUserRestaurantWaiter :one
-- Check if a user is a waiter for a given restaurant
SELECT id, user_id, restaurant_id, created_at, updated_at
FROM management.restaurants_waiters
WHERE user_id = $1
  AND restaurant_id = $2;
