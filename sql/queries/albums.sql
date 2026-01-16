-- name: CreateAlbum :one
INSERT INTO albums (title, description)
VALUES (?, ?)
RETURNING *;

-- name: GetAlbum :one
SELECT * FROM albums WHERE id = ?;

-- name: ListAlbums :many
SELECT * FROM albums
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAlbum :exec
UPDATE albums
SET title = ?, description = ?, cover_photo_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteAlbum :exec
DELETE FROM albums WHERE id = ?;
