CREATE SCHEMA IF NOT EXISTS management;

CREATE TABLE management.restaurants (
    id UUID PRIMARY KEY,
    name VARCHAR(150) UNIQUE NOT NULL,
    address VARCHAR(255) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
)
