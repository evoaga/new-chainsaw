// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user_providers.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const upsertUserProvider = `-- name: UpsertUserProvider :exec

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
    updated_at = EXCLUDED.updated_at
`

type UpsertUserProviderParams struct {
	UserID         int32              `json:"user_id"`
	Provider       string             `json:"provider"`
	ProviderUserID string             `json:"provider_user_id"`
	FirstName      pgtype.Text        `json:"first_name"`
	LastName       pgtype.Text        `json:"last_name"`
	Nickname       pgtype.Text        `json:"nickname"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Location       pgtype.Text        `json:"location"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

// User providers queries
func (q *Queries) UpsertUserProvider(ctx context.Context, arg UpsertUserProviderParams) error {
	_, err := q.db.Exec(ctx, upsertUserProvider,
		arg.UserID,
		arg.Provider,
		arg.ProviderUserID,
		arg.FirstName,
		arg.LastName,
		arg.Nickname,
		arg.AvatarUrl,
		arg.Location,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}
