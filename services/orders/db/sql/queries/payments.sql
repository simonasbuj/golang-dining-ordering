-- name: SavePayment :one
INSERT INTO orders.payments (
    id,
    order_id,
    amount_in_cents,
    currency,
    provider,
    provider_payment_id
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
