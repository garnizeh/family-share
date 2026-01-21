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

-- name: ListAlbumsWithPhotoCount :many
SELECT 
    a.id,
    a.title,
    a.description,
    a.cover_photo_id,
    a.created_at,
    a.updated_at,
    COUNT(p.id) as photo_count
FROM albums a
LEFT JOIN photos p ON p.album_id = a.id
GROUP BY a.id
ORDER BY a.created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAlbum :exec
UPDATE albums
SET title = ?, description = ?, cover_photo_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteAlbum :exec
DELETE FROM albums WHERE id = ?;

-- name: CountAlbums :one
SELECT COUNT(*) FROM albums;

-- name: SetAlbumCover :exec
UPDATE albums
SET cover_photo_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetAlbumWithPhotoCount :one
SELECT 
    a.*,
    COUNT(p.id) as photo_count
FROM albums a
LEFT JOIN photos p ON p.album_id = a.id
WHERE a.id = ?
GROUP BY a.id;
