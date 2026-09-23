-- name: CreatePlaylist :one
INSERT INTO playlists (
    id,
    owner_id,
    name,
    visibility,
    tracks,
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
    $7
)
RETURNING
    id,
    owner_id,
    name,
    visibility,
    tracks,
    created_at,
    updated_at;

-- name: GetPlaylistByID :one
SELECT
    id,
    owner_id,
    name,
    visibility,
    tracks,
    created_at,
    updated_at
FROM playlists
WHERE id = $1;

-- name: UpdatePlaylist :exec
UPDATE playlists
SET
    name = $2,
    visibility = $3,
    tracks = $4,
    updated_at = NOW()
WHERE id = $1;

-- name: DeletePlaylist :exec
DELETE FROM playlists
WHERE id = $1;

-- name: AddTrackToPlaylist :exec
UPDATE playlists
SET
    tracks = array_append(tracks, sqlc.arg(track_id)),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND NOT (sqlc.arg(track_id) = ANY(tracks));

-- name: RemoveTrackFromPlaylist :exec
UPDATE playlists
SET
    tracks = array_remove(tracks, sqlc.arg(track_id)),
    updated_at = NOW()
WHERE id = sqlc.arg(id);
