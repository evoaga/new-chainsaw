// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ExerciseType string

const (
	ExerciseTypeBodyweight ExerciseType = "Bodyweight"
	ExerciseTypeWeighted   ExerciseType = "Weighted"
	ExerciseTypeAssisted   ExerciseType = "Assisted"
)

func (e *ExerciseType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ExerciseType(s)
	case string:
		*e = ExerciseType(s)
	default:
		return fmt.Errorf("unsupported scan type for ExerciseType: %T", src)
	}
	return nil
}

type NullExerciseType struct {
	ExerciseType ExerciseType `json:"exercise_type"`
	Valid        bool         `json:"valid"` // Valid is true if ExerciseType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullExerciseType) Scan(value interface{}) error {
	if value == nil {
		ns.ExerciseType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ExerciseType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullExerciseType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ExerciseType), nil
}

type UnitSystem string

const (
	UnitSystemMetric   UnitSystem = "metric"
	UnitSystemImperial UnitSystem = "imperial"
)

func (e *UnitSystem) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UnitSystem(s)
	case string:
		*e = UnitSystem(s)
	default:
		return fmt.Errorf("unsupported scan type for UnitSystem: %T", src)
	}
	return nil
}

type NullUnitSystem struct {
	UnitSystem UnitSystem `json:"unit_system"`
	Valid      bool       `json:"valid"` // Valid is true if UnitSystem is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUnitSystem) Scan(value interface{}) error {
	if value == nil {
		ns.UnitSystem, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UnitSystem.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUnitSystem) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UnitSystem), nil
}

type BodyweightLog struct {
	ID         int32              `json:"id"`
	UserID     int32              `json:"user_id"`
	Bodyweight pgtype.Numeric     `json:"bodyweight"`
	LogDate    pgtype.Timestamptz `json:"log_date"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
}

type Exercise struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type ExerciseLog struct {
	ID               int32              `json:"id"`
	UserID           int32              `json:"user_id"`
	ExerciseID       int32              `json:"exercise_id"`
	Reps             int32              `json:"reps"`
	Weight           pgtype.Numeric     `json:"weight"`
	AdditionalWeight pgtype.Numeric     `json:"additional_weight"`
	ExerciseType     NullExerciseType   `json:"exercise_type"`
	BodyweightID     int32              `json:"bodyweight_id"`
	LogDate          pgtype.Timestamptz `json:"log_date"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
}

type InitialUserProvider struct {
	ID             int32              `json:"id"`
	UserID         int32              `json:"user_id"`
	Provider       string             `json:"provider"`
	ProviderUserID string             `json:"provider_user_id"`
	FirstName      pgtype.Text        `json:"first_name"`
	LastName       pgtype.Text        `json:"last_name"`
	Nickname       pgtype.Text        `json:"nickname"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Location       pgtype.Text        `json:"location"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
}

type RefreshToken struct {
	ID        int32              `json:"id"`
	UserID    int32              `json:"user_id"`
	Token     string             `json:"token"`
	ExpiresAt pgtype.Timestamp   `json:"expires_at"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type Trophy struct {
	ID          int32              `json:"id"`
	Name        string             `json:"name"`
	Description pgtype.Text        `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

type User struct {
	ID             int32              `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	Name           pgtype.Text        `json:"name"`
	Sex            pgtype.Text        `json:"sex"`
	PreferredUnits UnitSystem         `json:"preferred_units"`
	CountryCode    pgtype.Text        `json:"country_code"`
	AvatarUrl      pgtype.Text        `json:"avatar_url"`
	Bio            pgtype.Text        `json:"bio"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}

type UserProvider struct {
	ID             int32              `json:"id"`
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

type UserTrophy struct {
	ID           int32              `json:"id"`
	UserID       int32              `json:"user_id"`
	TrophyID     int32              `json:"trophy_id"`
	DisplayOrder pgtype.Int4        `json:"display_order"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}
