CREATE TABLE management.restaurant_managers (
    id VARCHAR(40) PRIMARY KEY,
    user_id VARCHAR(40) NOT NULL,
    restaurant_id VARCHAR(40) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_manager_user FOREIGN KEY (user_id)
        REFERENCES auth.users (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_manager_restaurant FOREIGN KEY (restaurant_id)
        REFERENCES management.restaurants (id)
        ON DELETE CASCADE,

    CONSTRAINT uq_manager UNIQUE (user_id, restaurant_id)
);
