CREATE TABLE management.tables (
    id UUID PRIMARY KEY,
    restaurant_id UUID NOT NULL,
    name VARCHAR(20) NOT NULL,
    capacity INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_table_restaurant FOREIGN KEY (restaurant_id)
        REFERENCES management.restaurants (id)
        ON DELETE SET NULL
);
