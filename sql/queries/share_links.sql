-- name: CreateShareLink :one
INSERT INTO share_links (token, target_type, target_id, max_views, expires_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetShareLinkByToken :one
SELECT * FROM share_links WHERE token = ?;

-- name: IncrementShareLinkView :exec
INSERT INTO share_link_views (share_link_id, viewer_hash) VALUES (?, ?);
