-- name: CreateActivityEvent :exec
INSERT INTO activity_events (event_type, album_id, photo_id, share_link_id) VALUES (?, ?, ?, ?);

-- name: ListRecentActivity :many
SELECT * FROM activity_events ORDER BY created_at DESC LIMIT ? OFFSET ?;
