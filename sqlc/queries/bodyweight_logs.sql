-- Body weight queries

-- name: LogBodyWeight :one
INSERT INTO bodyweight_logs (user_id, bodyweight, log_date)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetBodyWeightLogs :many
SELECT id, user_id, bodyweight, log_date, created_at, updated_at
FROM bodyweight_logs
WHERE user_id = $1
ORDER BY log_date DESC;

-- name: UpdateBodyWeight :exec
UPDATE bodyweight_logs
SET bodyweight = $2, updated_at = NOW()
WHERE user_id = $1 AND log_date = $3;

-- name: GetLatestBodyWeight :one
SELECT bodyweight
FROM bodyweight_logs
WHERE user_id = $1
ORDER BY log_date DESC
LIMIT 1;

-- name: GetBodyweightLogByUserIDAndDate :one
SELECT id
FROM bodyweight_logs
WHERE user_id = $1 AND log_date = $2;
