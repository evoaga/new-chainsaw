-- Refresh token queries

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE user_id = $1 AND token = $2;

-- name: DeleteAllRefreshTokensForUser :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: GetUserByRefreshToken :one
SELECT u.id, u.username, u.email, u.name
FROM users u
         JOIN refresh_tokens rt ON u.id = rt.user_id
WHERE rt.token = $1 AND rt.expires_at > CURRENT_TIMESTAMP;

-- name: InsertOrUpdateRefreshToken :exec
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at, created_at = CURRENT_TIMESTAMP;

-- name: ValidateRefreshToken :one
SELECT u.id, u.username, u.email, u.avatar_url, u.name
FROM users u
JOIN refresh_tokens rt ON u.id = rt.user_id
WHERE rt.token = $1 AND rt.expires_at > CURRENT_TIMESTAMP;
