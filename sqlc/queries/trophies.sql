-- Trophy queries

-- name: InsertUserTrophy :exec
INSERT INTO user_trophies (user_id, trophy_id, display_order)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, trophy_id) DO UPDATE
    SET display_order = EXCLUDED.display_order;

-- name: GetUserTrophies :many
SELECT t.id, t.name, t.description, ut.display_order
FROM trophies t
         JOIN user_trophies ut ON t.id = ut.trophy_id
WHERE ut.user_id = $1
ORDER BY ut.display_order, t.id;

-- name: GetTrophyByDisplayOrder :one
SELECT t.id, t.name, t.description, ut.display_order
FROM trophies t
         JOIN user_trophies ut ON t.id = ut.trophy_id
WHERE ut.user_id = $1 AND ut.display_order = $2;

-- name: DeleteUserTrophy :exec
DELETE FROM user_trophies
WHERE user_id = $1 AND trophy_id = $2;

-- name: DeleteUserTrophyByOrder :exec
DELETE FROM user_trophies
WHERE user_id = $1 AND display_order = $2;
