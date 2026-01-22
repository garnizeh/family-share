-- name: CreateShareLink :one
INSERT INTO share_links (token, target_type, target_id, max_views, expires_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetShareLinkByToken :one
SELECT * FROM share_links WHERE token = ?;

-- name: GetShareLink :one
SELECT * FROM share_links WHERE id = ?;

-- name: ListShareLinks :many
SELECT * FROM share_links
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListShareLinksWithDetails :many
SELECT 
    sl.*,
    CASE 
        WHEN sl.target_type = 'album' THEN a.title
        WHEN sl.target_type = 'photo' THEN (SELECT title FROM albums WHERE id = p.album_id)
    END as target_title,
    CASE
        WHEN sl.target_type = 'photo' THEN p.album_id
        ELSE NULL
    END as photo_album_id
FROM share_links sl
LEFT JOIN albums a ON sl.target_type = 'album' AND sl.target_id = a.id
LEFT JOIN photos p ON sl.target_type = 'photo' AND sl.target_id = p.id
ORDER BY sl.created_at DESC
LIMIT ? OFFSET ?;

-- name: ListActiveShareLinks :many
SELECT * FROM share_links
WHERE revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: RevokeShareLink :exec
UPDATE share_links
SET revoked_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: CountShareLinks :one
SELECT COUNT(*) FROM share_links;

-- name: IncrementShareLinkView :exec
INSERT INTO share_link_views (share_link_id, viewer_hash) VALUES (?, ?);
