CREATE SCHEMA IF NOT EXISTS orders;

CREATE TYPE order_status AS ENUM (
    'open',
    'locked',
    'completed',
    'cancelled'
);

CREATE TABLE orders.orders (
    id UUID PRIMARY KEY,
    table_id UUID NOT NULL,
    status order_status NOT NULL DEFAULT 'open',
    currency varchar(3) NOT NULL,
    tip_amount_in_cents INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_table FOREIGN KEY (table_id)
        REFERENCES management.tables (id)
        ON DELETE CASCADE
);



CREATE TABLE orders.orders_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    item_id UUID,
    item_name VARCHAR(40) NOT NULL,
    price_in_cents INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_items_order FOREIGN KEY (order_id)
        REFERENCES orders.orders (id)
        ON DELETE CASCADE,
    
    CONSTRAINT fk_orders_items_item FOREIGN KEY (item_id)
        REFERENCES management.items (id)
        ON DELETE SET NULL
);
