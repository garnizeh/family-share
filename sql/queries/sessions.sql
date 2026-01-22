-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = ?;
