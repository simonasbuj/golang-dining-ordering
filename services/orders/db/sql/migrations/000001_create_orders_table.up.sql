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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_table FOREIGN KEY (table_id)
        REFERENCES management.tables (id)
        ON DELETE CASCADE
);
