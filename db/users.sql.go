// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteUser = `-- name: DeleteUser :exec

DELETE FROM users WHERE id = $1
`

// User queries
func (q *Queries) DeleteUser(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE email = $1
`

type GetUserByEmailRow struct {
	ID             int32              `json:"id"`
	Username       string             `json:"username"`
	Name           pgtype.Text        `json:"name"`
	Email          string             `json:"email"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Sex            pgtype.Text        `json:"sex"`
	PreferredUnits UnitSystem         `json:"preferred_units"`
	CountryCode    pgtype.Text        `json:"country_code"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Name,
		&i.Email,
		&i.AvatarUrl,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE id = $1
`

type GetUserByIDRow struct {
	ID             int32              `json:"id"`
	Username       string             `json:"username"`
	Name           pgtype.Text        `json:"name"`
	Email          string             `json:"email"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Sex            pgtype.Text        `json:"sex"`
	PreferredUnits UnitSystem         `json:"preferred_units"`
	CountryCode    pgtype.Text        `json:"country_code"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserByID(ctx context.Context, id int32) (GetUserByIDRow, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i GetUserByIDRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Name,
		&i.Email,
		&i.AvatarUrl,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, name, email, avatar_url, sex, preferred_units, country_code, created_at, updated_at
FROM users
WHERE username = $1
`

type GetUserByUsernameRow struct {
	ID             int32              `json:"id"`
	Username       string             `json:"username"`
	Name           pgtype.Text        `json:"name"`
	Email          string             `json:"email"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Sex            pgtype.Text        `json:"sex"`
	PreferredUnits UnitSystem         `json:"preferred_units"`
	CountryCode    pgtype.Text        `json:"country_code"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (GetUserByUsernameRow, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i GetUserByUsernameRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Name,
		&i.Email,
		&i.AvatarUrl,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserDetails = `-- name: GetUserDetails :one
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
    (SELECT DISTINCT ON (user_id) id, user_id, bodyweight, log_date, created_at, updated_at FROM bodyweight_logs ORDER BY user_id, log_date DESC) bw
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
    u.id, bw.bodyweight
`

type GetUserDetailsRow struct {
	UserID           int32          `json:"user_id"`
	Username         string         `json:"username"`
	Name             pgtype.Text    `json:"name"`
	Email            string         `json:"email"`
	AvatarUrl        pgtype.Text    `json:"avatar_url"`
	Sex              pgtype.Text    `json:"sex"`
	PreferredUnits   UnitSystem     `json:"preferred_units"`
	CountryCode      pgtype.Text    `json:"country_code"`
	LatestBodyWeight pgtype.Numeric `json:"latest_body_weight"`
	ExerciseLogs     interface{}    `json:"exercise_logs"`
}

func (q *Queries) GetUserDetails(ctx context.Context, id int32) (GetUserDetailsRow, error) {
	row := q.db.QueryRow(ctx, getUserDetails, id)
	var i GetUserDetailsRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.Name,
		&i.Email,
		&i.AvatarUrl,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.LatestBodyWeight,
		&i.ExerciseLogs,
	)
	return i, err
}

const getUserPreferences = `-- name: GetUserPreferences :one
SELECT sex, preferred_units
FROM users
WHERE id = $1
`

type GetUserPreferencesRow struct {
	Sex            pgtype.Text `json:"sex"`
	PreferredUnits UnitSystem  `json:"preferred_units"`
}

func (q *Queries) GetUserPreferences(ctx context.Context, id int32) (GetUserPreferencesRow, error) {
	row := q.db.QueryRow(ctx, getUserPreferences, id)
	var i GetUserPreferencesRow
	err := row.Scan(&i.Sex, &i.PreferredUnits)
	return i, err
}

const getUserPreferredUnit = `-- name: GetUserPreferredUnit :one
SELECT preferred_units FROM users WHERE id = $1
`

func (q *Queries) GetUserPreferredUnit(ctx context.Context, id int32) (UnitSystem, error) {
	row := q.db.QueryRow(ctx, getUserPreferredUnit, id)
	var preferred_units UnitSystem
	err := row.Scan(&preferred_units)
	return preferred_units, err
}

const getUserProfileByID = `-- name: GetUserProfileByID :one
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
WHERE u.id = $1
`

type GetUserProfileByIDRow struct {
	ID                 int32       `json:"id"`
	Username           string      `json:"username"`
	Name               pgtype.Text `json:"name"`
	Sex                pgtype.Text `json:"sex"`
	PreferredUnits     UnitSystem  `json:"preferred_units"`
	CountryCode        pgtype.Text `json:"country_code"`
	AvatarUrl          pgtype.Text `json:"avatar_url"`
	Bio                pgtype.Text `json:"bio"`
	ExerciseLogs       interface{} `json:"exercise_logs"`
	Trophies           interface{} `json:"trophies"`
	BodyweightLogs     interface{} `json:"bodyweight_logs"`
	ExerciseLogsRecent interface{} `json:"exercise_logs_recent"`
	LatestBodyweight   interface{} `json:"latest_bodyweight"`
}

func (q *Queries) GetUserProfileByID(ctx context.Context, id int32) (GetUserProfileByIDRow, error) {
	row := q.db.QueryRow(ctx, getUserProfileByID, id)
	var i GetUserProfileByIDRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Name,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.AvatarUrl,
		&i.Bio,
		&i.ExerciseLogs,
		&i.Trophies,
		&i.BodyweightLogs,
		&i.ExerciseLogsRecent,
		&i.LatestBodyweight,
	)
	return i, err
}

const getUserProfileByUsername = `-- name: GetUserProfileByUsername :one
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
WHERE u.username = $1
`

type GetUserProfileByUsernameRow struct {
	ID                 int32       `json:"id"`
	Username           string      `json:"username"`
	Name               pgtype.Text `json:"name"`
	Sex                pgtype.Text `json:"sex"`
	PreferredUnits     UnitSystem  `json:"preferred_units"`
	CountryCode        pgtype.Text `json:"country_code"`
	AvatarUrl          pgtype.Text `json:"avatar_url"`
	Bio                pgtype.Text `json:"bio"`
	ExerciseLogs       interface{} `json:"exercise_logs"`
	Trophies           interface{} `json:"trophies"`
	BodyweightLogs     interface{} `json:"bodyweight_logs"`
	ExerciseLogsRecent interface{} `json:"exercise_logs_recent"`
	LatestBodyweight   interface{} `json:"latest_bodyweight"`
}

func (q *Queries) GetUserProfileByUsername(ctx context.Context, username string) (GetUserProfileByUsernameRow, error) {
	row := q.db.QueryRow(ctx, getUserProfileByUsername, username)
	var i GetUserProfileByUsernameRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Name,
		&i.Sex,
		&i.PreferredUnits,
		&i.CountryCode,
		&i.AvatarUrl,
		&i.Bio,
		&i.ExerciseLogs,
		&i.Trophies,
		&i.BodyweightLogs,
		&i.ExerciseLogsRecent,
		&i.LatestBodyweight,
	)
	return i, err
}

const getUserSex = `-- name: GetUserSex :one
SELECT sex
FROM users
WHERE id = $1
`

func (q *Queries) GetUserSex(ctx context.Context, id int32) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, getUserSex, id)
	var sex pgtype.Text
	err := row.Scan(&sex)
	return sex, err
}

const insertUser = `-- name: InsertUser :one
INSERT INTO users (email, username, name, avatar_url, sex, preferred_units, country_code, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, username
`

type InsertUserParams struct {
	Email          string             `json:"email"`
	Username       string             `json:"username"`
	Name           pgtype.Text        `json:"name"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Sex            pgtype.Text        `json:"sex"`
	PreferredUnits UnitSystem         `json:"preferred_units"`
	CountryCode    pgtype.Text        `json:"country_code"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

type InsertUserRow struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
}

func (q *Queries) InsertUser(ctx context.Context, arg InsertUserParams) (InsertUserRow, error) {
	row := q.db.QueryRow(ctx, insertUser,
		arg.Email,
		arg.Username,
		arg.Name,
		arg.AvatarUrl,
		arg.Sex,
		arg.PreferredUnits,
		arg.CountryCode,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i InsertUserRow
	err := row.Scan(&i.ID, &i.Username)
	return i, err
}

const searchUsers = `-- name: SearchUsers :many
SELECT username, name, avatar_url, sex, country_code
FROM users
WHERE
    username ILIKE '%' || $1 || '%'
   OR name ILIKE '%' || $1 || '%'
ORDER BY username
LIMIT $2 OFFSET $3
`

type SearchUsersParams struct {
	Column1 pgtype.Text `json:"column_1"`
	Limit   int32       `json:"limit"`
	Offset  int32       `json:"offset"`
}

type SearchUsersRow struct {
	Username    string      `json:"username"`
	Name        pgtype.Text `json:"name"`
	AvatarUrl   pgtype.Text `json:"avatar_url"`
	Sex         pgtype.Text `json:"sex"`
	CountryCode pgtype.Text `json:"country_code"`
}

func (q *Queries) SearchUsers(ctx context.Context, arg SearchUsersParams) ([]SearchUsersRow, error) {
	rows, err := q.db.Query(ctx, searchUsers, arg.Column1, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchUsersRow
	for rows.Next() {
		var i SearchUsersRow
		if err := rows.Scan(
			&i.Username,
			&i.Name,
			&i.AvatarUrl,
			&i.Sex,
			&i.CountryCode,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUser = `-- name: UpdateUser :exec
UPDATE users
SET username = $2,
    name = $3,
    avatar_url = $4,
    sex = $5,
    preferred_units = $6,
    country_code = $7,
    updated_at = NOW()
WHERE id = $1
`

type UpdateUserParams struct {
	ID             int32       `json:"id"`
	Username       string      `json:"username"`
	Name           pgtype.Text `json:"name"`
	AvatarUrl      pgtype.Text `json:"avatar_url"`
	Sex            pgtype.Text `json:"sex"`
	PreferredUnits UnitSystem  `json:"preferred_units"`
	CountryCode    pgtype.Text `json:"country_code"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) error {
	_, err := q.db.Exec(ctx, updateUser,
		arg.ID,
		arg.Username,
		arg.Name,
		arg.AvatarUrl,
		arg.Sex,
		arg.PreferredUnits,
		arg.CountryCode,
	)
	return err
}

const updateUsername = `-- name: UpdateUsername :exec
UPDATE users SET username = $1, updated_at = NOW() WHERE id = $2
`

type UpdateUsernameParams struct {
	Username string `json:"username"`
	ID       int32  `json:"id"`
}

func (q *Queries) UpdateUsername(ctx context.Context, arg UpdateUsernameParams) error {
	_, err := q.db.Exec(ctx, updateUsername, arg.Username, arg.ID)
	return err
}
