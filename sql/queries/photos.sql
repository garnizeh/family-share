-- name: CreatePhoto :one
INSERT INTO photos (album_id, filename, width, height, size_bytes, format)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPhoto :one
SELECT * FROM photos WHERE id = ?;

-- name: ListPhotosByAlbum :many
SELECT * FROM photos WHERE album_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: DeletePhoto :exec
DELETE FROM photos WHERE id = ?;
