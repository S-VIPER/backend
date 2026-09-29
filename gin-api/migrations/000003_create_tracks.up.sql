CREATE TABLE tracks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    album_title TEXT NOT NULL DEFAULT '',
    album_art_url TEXT NOT NULL DEFAULT '',
    genre TEXT[] NOT NULL DEFAULT '{}'::text[],
    year INTEGER NOT NULL DEFAULT 0 CHECK (year >= 0 AND year <= 2100),
    musicbrainz_release_id TEXT NOT NULL DEFAULT '',
    musicbrainz_release_group_id TEXT NOT NULL DEFAULT '',
    object_key TEXT NOT NULL UNIQUE,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX tracks_artist_idx ON tracks (artist);
CREATE INDEX tracks_title_idx ON tracks (title);
CREATE INDEX tracks_release_id_idx ON tracks (musicbrainz_release_id);
