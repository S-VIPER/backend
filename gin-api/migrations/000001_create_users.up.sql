CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    email_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_not_empty
        CHECK (length(trim(email)) > 0),

    CONSTRAINT users_email_length
        CHECK (length(email) <= 320),

    CONSTRAINT users_password_hash_not_empty
        CHECK (length(trim(password_hash)) > 0)
);

CREATE UNIQUE INDEX users_email_unique
    ON users (lower(email));