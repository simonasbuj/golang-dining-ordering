CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(60) NOT NULL,
    name VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    role INT DEFAULT 2 NOT NULL,
    is_active Boolean DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS auth.tokens (
    id TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_token_user FOREIGN KEY (user_id)
        REFERENCES auth.users (id)
        ON DELETE CASCADE
);
