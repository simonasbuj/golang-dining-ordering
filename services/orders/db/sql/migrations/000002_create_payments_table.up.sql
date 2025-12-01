CREATE TYPE orders.payment_provider AS ENUM (
    'stripe',
    'mock',
    'klix'
);

CREATE TABLE orders.payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    amount_in_cents INT NOT NULL,
    currency varchar(3) NOT NULL,
    provider orders.payment_provider NOT NULL,
    provider_payment_id varchar(30) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    refunded_at TIMESTAMPTZ,

    CONSTRAINT fk_payment_order FOREIGN KEY (order_id)
        REFERENCES orders.orders (id)
        ON DELETE SET NULL
);
