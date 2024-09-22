-- User providers queries

-- name: UpsertUserProvider :exec
INSERT INTO user_providers (
    user_id, provider, provider_user_id, first_name, last_name, nickname, avatar_url, location, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) ON CONFLICT (user_id, provider, provider_user_id) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    nickname = EXCLUDED.nickname,
    avatar_url = EXCLUDED.avatar_url,
    location = EXCLUDED.location,
    updated_at = EXCLUDED.updated_at;
