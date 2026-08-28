CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    code_hash TEXT NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    attempts INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    verified_at TIMESTAMPTZ,

    CONSTRAINT email_verifications_attempts_non_negative
        CHECK (attempts >= 0),

    CONSTRAINT email_verifications_code_hash_not_empty
        CHECK (length(trim(code_hash)) > 0)
);

CREATE INDEX idx_email_verifications_user_id
    ON email_verifications(user_id);

CREATE INDEX idx_email_verifications_expires_at
    ON email_verifications(expires_at);