-- Initial user providers queries

-- name: InsertInitialUserProvider :exec
INSERT INTO initial_user_providers (user_id, provider, provider_user_id, first_name, last_name, nickname, avatar_url, location, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (user_id, provider) DO NOTHING;
