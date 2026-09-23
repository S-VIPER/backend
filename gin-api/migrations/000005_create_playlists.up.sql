CREATE TABLE playlists (
    id UUID PRIMARY KEY DEFAULT uuidv7(),

    owner_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    name TEXT NOT NULL,

    visibility TEXT NOT NULL DEFAULT 'private',

   tracks TEXT[] NOT NULL DEFAULT '{}'::text[],

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT playlists_name_not_empty
        CHECK (length(trim(name)) > 0),

    CONSTRAINT playlists_visibility_valid
        CHECK (visibility IN ('private', 'public'))
);

CREATE INDEX idx_playlists_owner_id
    ON playlists(owner_id);

CREATE INDEX idx_playlists_visibility
    ON playlists(visibility) WHERE visibility = 'public';
