CREATE TABLE orders.orders_waiters (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    order_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT fk_orderswaiters_users FOREIGN KEY (user_id)
        REFERENCES auth.users (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_orderswaiters_order FOREIGN KEY (order_id)
        REFERENCES orders.orders (id)
        ON DELETE CASCADE,

    CONSTRAINT uq_order_waiter UNIQUE (user_id, order_id)
);
