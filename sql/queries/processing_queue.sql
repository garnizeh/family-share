-- name: EnqueueJob :one
INSERT INTO processing_queue (
    album_id, original_filename, temp_filepath, status
) VALUES (
    ?, ?, ?, 'pending'
)
RETURNING *;

-- name: GetNextPendingJob :one
UPDATE processing_queue
SET status = 'processing', updated_at = CURRENT_TIMESTAMP
WHERE id = (
  SELECT id FROM processing_queue
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT 1
)
RETURNING *;

-- name: UpdateJobStatus :exec
UPDATE processing_queue
SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteJob :exec
DELETE FROM processing_queue
WHERE id = ?;

-- name: GetQueueStatus :one
SELECT 
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) AS pending_count,
    SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END) AS processing_count,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failed_count,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed_count 
FROM processing_queue
WHERE album_id = ?;

-- name: ListFailedJobs :many
SELECT * FROM processing_queue
WHERE album_id = ? AND status = 'failed'
ORDER BY created_at ASC;

-- name: ClearFailedJobs :exec
DELETE FROM processing_queue
WHERE album_id = ? AND status = 'failed';

-- name: CountActiveJobs :one
SELECT COUNT(*) FROM processing_queue 
WHERE album_id = ? AND status IN ('pending', 'processing');
