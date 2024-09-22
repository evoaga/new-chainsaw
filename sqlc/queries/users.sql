-- User queries

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE username = $1;

-- name: UpdateUsername :exec
UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUser :exec
UPDATE users
SET username = $2,
    name = $3,
    avatar_url = $4,
    sex = $5,
    preferred_units = $6,
    country_code = $7,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE email = $1;

-- name: InsertUser :one
INSERT INTO users (email, username, name, avatar_url, sex, preferred_units, country_code, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, username;

-- name: GetUserByID :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserSex :one
SELECT sex
FROM users
WHERE id = $1;

-- name: GetUserPreferredUnit :one
SELECT preferred_units FROM users WHERE id = $1;

-- name: GetUserDetails :one
SELECT
    u.id as user_id,
    u.username,
    u.name,
    u.email,
    u.avatar_url,
    u.sex,
    u.preferred_units,
    u.country_code,
    bw.bodyweight as latest_body_weight,
    COALESCE(json_agg(json_build_object(
        'exercise_id', el.exercise_id,
        'exercise_name', e.name,
        'reps', el.reps,
        'weight', el.weight,
        'log_date', el.log_date
    )) FILTER (WHERE el.id IS NOT NULL), '[]') as exercise_logs
FROM
    users u
LEFT JOIN
    (SELECT DISTINCT ON (user_id) * FROM bodyweight_logs ORDER BY user_id, log_date DESC) bw
    ON u.id = bw.user_id
LEFT JOIN
    exercise_logs el
    ON u.id = el.user_id
LEFT JOIN
    exercises e
    ON el.exercise_id = e.id
WHERE
    u.id = $1
GROUP BY
    u.id, bw.bodyweight;

-- name: GetUserProfileByUsername :one
SELECT
    u.id,
    u.username,
    u.name,
    u.sex,
    u.preferred_units,
    u.country_code,
    u.avatar_url,
    u.bio,
    (
        SELECT COALESCE(json_agg(el), '[]'::json)
        FROM (
                 SELECT
                     el.id,
                     el.exercise_id,
                     e.name AS exercise_name,
                     el.reps,
                     el.weight,
                     el.additional_weight,
                     el.exercise_type,
                     el.log_date,
                     bw.bodyweight
                 FROM exercise_logs el
                          JOIN exercises e ON el.exercise_id = e.id
                          JOIN bodyweight_logs bw ON el.bodyweight_id = bw.id
                 WHERE el.user_id = u.id
                 ORDER BY el.log_date DESC
             ) el
    ) AS exercise_logs,
    (
        SELECT COALESCE(json_agg(t), '[]'::json)
        FROM (
                 SELECT
                     t.id,
                     t.name,
                     t.description,
                     ut.display_order
                 FROM user_trophies ut
                          JOIN trophies t ON ut.trophy_id = t.id
                 WHERE ut.user_id = u.id
             ) t
    ) AS trophies,
    (
        SELECT COALESCE(json_agg(bw), '[]'::json)
        FROM (
                 SELECT
                     bw.id,
                     bw.bodyweight,
                     bw.log_date,
                     bw.created_at,
                     bw.updated_at
                 FROM bodyweight_logs bw
                 WHERE bw.user_id = u.id
                 ORDER BY bw.log_date DESC
             ) bw
    ) AS bodyweight_logs,
    (
        SELECT COALESCE(json_agg(el), '[]'::json)
        FROM (
                 SELECT
                     el.id,
                     el.exercise_id,
                     e.name AS exercise_name,
                     el.reps,
                     el.weight,
                     el.additional_weight,
                     el.exercise_type,
                     el.log_date,
                     bw.bodyweight
                 FROM exercise_logs el
                          JOIN exercises e ON el.exercise_id = e.id
                          JOIN bodyweight_logs bw ON el.bodyweight_id = bw.id
                 WHERE el.user_id = u.id
                   AND el.log_date = (
                     SELECT MAX(sub_el.log_date)
                     FROM exercise_logs sub_el
                     WHERE sub_el.user_id = el.user_id
                       AND sub_el.exercise_id = el.exercise_id
                 )
                 ORDER BY el.log_date DESC
             ) el
    ) AS exercise_logs_recent,
    (
        SELECT COALESCE(jsonb_build_object(
            'id', bw.id,
            'bodyweight', bw.bodyweight,
            'log_date', bw.log_date,
            'created_at', bw.created_at,
            'updated_at', bw.updated_at
        ), '{}'::jsonb)
        FROM (
            SELECT
                bw.id,
                bw.bodyweight,
                bw.log_date,
                bw.created_at,
                bw.updated_at
            FROM bodyweight_logs bw
            WHERE bw.user_id = u.id
            ORDER BY bw.log_date DESC
            LIMIT 1
        ) bw
    ) AS latest_bodyweight
FROM users u
WHERE u.username = $1;

-- name: GetUserProfileByID :one
SELECT
    u.id,
    u.username,
    u.name,
    u.sex,
    u.preferred_units,
    u.country_code,
    u.avatar_url,
    u.bio,
    (
        SELECT COALESCE(json_agg(el), '[]'::json)
        FROM (
                 SELECT
                     el.id,
                     el.exercise_id,
                     e.name AS exercise_name,
                     el.reps,
                     el.weight,
                     el.additional_weight,
                     el.exercise_type,
                     el.log_date,
                     bw.bodyweight
                 FROM exercise_logs el
                          JOIN exercises e ON el.exercise_id = e.id
                          JOIN bodyweight_logs bw ON el.bodyweight_id = bw.id
                 WHERE el.user_id = u.id
                 ORDER BY el.log_date DESC
             ) el
    ) AS exercise_logs,
    (
        SELECT COALESCE(json_agg(t), '[]'::json)
        FROM (
                 SELECT
                     t.id,
                     t.name,
                     t.description,
                     ut.display_order
                 FROM user_trophies ut
                          JOIN trophies t ON ut.trophy_id = t.id
                 WHERE ut.user_id = u.id
             ) t
    ) AS trophies,
    (
        SELECT COALESCE(json_agg(bw), '[]'::json)
        FROM (
                 SELECT
                     bw.id,
                     bw.bodyweight,
                     bw.log_date,
                     bw.created_at,
                     bw.updated_at
                 FROM bodyweight_logs bw
                 WHERE bw.user_id = u.id
                 ORDER BY bw.log_date DESC
             ) bw
    ) AS bodyweight_logs,
    (
        SELECT COALESCE(json_agg(el), '[]'::json)
        FROM (
                 SELECT
                     el.id,
                     el.exercise_id,
                     e.name AS exercise_name,
                     el.reps,
                     el.weight,
                     el.additional_weight,
                     el.exercise_type,
                     el.log_date,
                     bw.bodyweight
                 FROM exercise_logs el
                          JOIN exercises e ON el.exercise_id = e.id
                          JOIN bodyweight_logs bw ON el.bodyweight_id = bw.id
                 WHERE el.user_id = u.id
                   AND el.log_date = (
                     SELECT MAX(sub_el.log_date)
                     FROM exercise_logs sub_el
                     WHERE sub_el.user_id = el.user_id
                       AND sub_el.exercise_id = el.exercise_id
                 )
                 ORDER BY el.log_date DESC
             ) el
    ) AS exercise_logs_recent,
    (
        SELECT COALESCE(jsonb_build_object(
            'id', bw.id,
            'bodyweight', bw.bodyweight,
            'log_date', bw.log_date,
            'created_at', bw.created_at,
            'updated_at', bw.updated_at
        ), '{}'::jsonb)
        FROM (
            SELECT
                bw.id,
                bw.bodyweight,
                bw.log_date,
                bw.created_at,
                bw.updated_at
            FROM bodyweight_logs bw
            WHERE bw.user_id = u.id
            ORDER BY bw.log_date DESC
            LIMIT 1
        ) bw
    ) AS latest_bodyweight
FROM users u
WHERE u.id = $1;

-- name: GetUserPreferences :one
SELECT sex, preferred_units
FROM users
WHERE id = $1;

-- name: SearchUsers :many
SELECT username, name, avatar_url, sex, country_code
FROM users
WHERE
    username ILIKE '%' || $1 || '%'
   OR name ILIKE '%' || $1 || '%'
ORDER BY username
LIMIT $2 OFFSET $3;