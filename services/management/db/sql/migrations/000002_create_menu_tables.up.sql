CREATE TABLE management.menus (
    id UUID PRIMARY KEY,
    restaurant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_menu_restaurant FOREIGN KEY (restaurant_id)
        REFERENCES management.restaurants (id)
        ON DELETE SET NULL,

    CONSTRAINT uq_menu UNIQUE (id, restaurant_id)
);

CREATE TABLE management.categories (
    id UUID PRIMARY KEY,
    menu_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_category_menu FOREIGN KEY (menu_id)
        REFERENCES management.menus (id)
        ON DELETE SET NULL,

    CONSTRAINT uq_category UNIQUE (menu_id, name)
);

CREATE TABLE management.items (
    id UUID PRIMARY KEY,
    category_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    price_in_cents INTEGER NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT True,
    image_path VARCHAR(200),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_item_category FOREIGN KEY (category_id)
        REFERENCES management.categories (id)
        ON DELETE SET NULL,

    CONSTRAINT uq_item UNIQUE (category_id, name)
);
