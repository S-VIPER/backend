-- name: CreateTrack :one
INSERT INTO tracks (
    id,
    title,
    artist,
    album_title,
    album_art_url,
    genre,
    year,
    musicbrainz_release_id,
    musicbrainz_release_group_id,
    object_key,
    file_name,
    content_type,
    file_size,
    created_at,
    updated_at
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15
)
RETURNING
    id,
    title,
    artist,
    album_title,
    album_art_url,
    genre,
    year,
    musicbrainz_release_id,
    musicbrainz_release_group_id,
    object_key,
    file_name,
    content_type,
    file_size,
    created_at,
    updated_at;

-- name: GetTrackByID :one
SELECT
    id,
    title,
    artist,
    album_title,
    album_art_url,
    genre,
    year,
    musicbrainz_release_id,
    musicbrainz_release_group_id,
    object_key,
    file_name,
    content_type,
    file_size,
    created_at,
    updated_at
FROM tracks
WHERE id = $1;

-- name: GetAllTracks :many
SELECT
    id,
    title,
    artist,
    album_title,
    album_art_url,
    genre,
    year,
    musicbrainz_release_id,
    musicbrainz_release_group_id,
    object_key,
    file_name,
    content_type,
    file_size,
    created_at,
    updated_at
FROM tracks
ORDER BY artist, title, id;

-- name: UpdateTrack :execrows
UPDATE tracks
SET
    title = $2,
    artist = $3,
    album_title = $4,
    album_art_url = $5,
    genre = $6,
    year = $7,
    musicbrainz_release_id = $8,
    musicbrainz_release_group_id = $9,
    updated_at = $10
WHERE id = $1;

-- name: DeleteTrack :execrows
DELETE FROM tracks
WHERE id = $1;

-- name: TrackExists :one
SELECT EXISTS (
    SELECT 1
    FROM tracks
    WHERE id = $1
);
