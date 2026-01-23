-- name: CreateActivityEvent :exec
INSERT INTO activity_events (event_type, album_id, photo_id, share_link_id) VALUES (?, ?, ?, ?);

-- name: ListRecentActivity :many
SELECT * FROM activity_events ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountActivityByTypeSince :many
SELECT event_type, COUNT(*) as count
FROM activity_events
WHERE created_at >= ?
GROUP BY event_type;

-- name: CountUploadsSince :one
SELECT COUNT(*) FROM activity_events
WHERE event_type = 'upload' AND created_at >= ?;

-- name: CountAlbumViewsSince :one
SELECT COUNT(*) FROM activity_events
WHERE event_type = 'album_view' AND created_at >= ?;

-- name: CountPhotoViewsSince :one
SELECT COUNT(*) FROM activity_events
WHERE event_type = 'photo_view' AND created_at >= ?;

-- name: CountShareViewsSince :one
SELECT COUNT(*) FROM activity_events
WHERE event_type = 'share_view' AND created_at >= ?;

-- name: DeleteOldActivityEvents :exec
DELETE FROM activity_events WHERE created_at < ?;
