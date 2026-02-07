-- name: CreatePhoto :one
INSERT INTO photos (album_id, filename, width, height, size_bytes, format)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPhoto :one
SELECT * FROM photos WHERE id = ?;

-- name: ListPhotosByAlbum :many
SELECT * FROM photos WHERE album_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: ListAllPhotosWithAlbum :many
SELECT 
    p.*,
    a.title as album_title
FROM photos p
JOIN albums a ON p.album_id = a.id
ORDER BY p.created_at DESC
LIMIT ? OFFSET ?;

-- name: DeletePhoto :exec
DELETE FROM photos WHERE id = ?;

-- name: CountPhotos :one
SELECT COUNT(*) FROM photos;

-- name: GetTotalStorageBytes :one
SELECT COALESCE(SUM(size_bytes), 0) FROM photos;

-- name: UpdatePhotoDimensions :exec
UPDATE photos
SET width = ?, height = ?, size_bytes = ?
WHERE id = ?;

-- name: DeleteOrphanedPhotos :many
DELETE FROM photos 
WHERE album_id NOT IN (SELECT id FROM albums)
RETURNING id, album_id, filename, format, created_at;
