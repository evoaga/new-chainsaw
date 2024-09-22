package main

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"
	"new-chainsaw/db"
)

var (
	database  = os.Getenv("DB_DATABASE")
	password  = os.Getenv("DB_PASSWORD")
	username  = os.Getenv("DB_USERNAME")
	port      = os.Getenv("DB_PORT")
	host      = os.Getenv("DB_HOST")
	startDate = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate   = time.Now()
)

func main() {
	// Initialize gofakeit
	err := gofakeit.Seed(0)
	if err != nil {
		return
	}

	// Build the connection string using environment variables
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}

	dbConn, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	ctx := context.Background()

	userIDs := insertUsers(ctx, dbConn, 10)
	insertUserProviders(ctx, dbConn, userIDs)
	insertRefreshTokens(ctx, dbConn, userIDs)
	insertExerciseLogs(ctx, dbConn, userIDs)
	insertUserTrophies(ctx, dbConn, userIDs)
	//insertBodyweightLogs(ctx, dbConn, userIDs)

	fmt.Println("Fake data insertion completed")
}

func randomUnitSystem() db.UnitSystem {
	if gofakeit.Bool() {
		return db.UnitSystemMetric
	}
	return db.UnitSystemImperial
}

// Function to generate a random date between a start date and end date
func randomDateBetween(start, end time.Time) time.Time {
	startUnix := start.Unix()
	endUnix := end.Unix()
	delta := endUnix - startUnix

	randomSec := rand.Int63n(delta) + startUnix
	return time.Unix(randomSec, 0)
}

func insertUsers(ctx context.Context, dbConn *pgxpool.Pool, count int) []int32 {
	userIDs := make([]int32, 0, count)
	for i := 0; i < count; i++ {
		user := db.User{
			Username:       gofakeit.Username(),
			Name:           pgtype.Text{String: gofakeit.Name(), Valid: true},
			Email:          gofakeit.Email(),
			AvatarUrl:      pgtype.Text{String: "https://lh3.googleusercontent.com/a-/ALV-UjWpT38p8X51UL9XOWOeGbuKb0Dqjun2krYv0f26CXZTpPnW=s96-c", Valid: true},
			Sex:            pgtype.Text{String: gofakeit.RandomString([]string{"male", "female"}), Valid: true},
			PreferredUnits: randomUnitSystem(),
			CountryCode:    pgtype.Text{String: gofakeit.CountryAbr(), Valid: true},
			Bio:            pgtype.Text{String: gofakeit.SentenceSimple(), Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		var userID int32
		err := dbConn.QueryRow(ctx, `
            INSERT INTO users (username, name, email, avatar_url, sex, preferred_units, country_code, bio, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
            RETURNING id`, user.Username, user.Name, user.Email, user.AvatarUrl, user.Sex, user.PreferredUnits, user.CountryCode, user.Bio, user.CreatedAt, user.UpdatedAt).Scan(&userID)
		if err != nil {
			log.Printf("Error inserting user %v: %v", user, err)
			continue
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func insertUserProviders(ctx context.Context, dbConn *pgxpool.Pool, userIDs []int32) {
	for _, userID := range userIDs {
		userProvider := db.UserProvider{
			UserID:         userID,
			Provider:       gofakeit.Company(),
			ProviderUserID: gofakeit.UUID(),
			FirstName:      pgtype.Text{String: gofakeit.FirstName(), Valid: true},
			LastName:       pgtype.Text{String: gofakeit.LastName(), Valid: true},
			Nickname:       pgtype.Text{String: gofakeit.Username(), Valid: true},
			AvatarUrl:      pgtype.Text{String: gofakeit.URL(), Valid: true},
			Location:       pgtype.Text{String: gofakeit.City(), Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		_, err := dbConn.Exec(ctx, `
            INSERT INTO user_providers (user_id, provider, provider_user_id, first_name, last_name, nickname, avatar_url, location, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, userProvider.UserID, userProvider.Provider, userProvider.ProviderUserID, userProvider.FirstName, userProvider.LastName, userProvider.Nickname, userProvider.AvatarUrl, userProvider.Location, userProvider.CreatedAt, userProvider.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting user provider %v: %v", userProvider, err)
		}
	}
}

func insertRefreshTokens(ctx context.Context, dbConn *pgxpool.Pool, userIDs []int32) {
	for _, userID := range userIDs {
		refreshToken := db.RefreshToken{
			UserID:    userID,
			Token:     gofakeit.UUID(),
			ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		_, err := dbConn.Exec(ctx, `
            INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
            VALUES ($1, $2, $3, $4)`, refreshToken.UserID, refreshToken.Token, refreshToken.ExpiresAt, refreshToken.CreatedAt)
		if err != nil {
			log.Printf("Error inserting refresh token %v: %v", refreshToken, err)
		}
	}
}

func insertExerciseLogs(ctx context.Context, dbConn *pgxpool.Pool, userIDs []int32) {
	exerciseIDs := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, userID := range userIDs {
		numExercises := 20
		if numExercises > len(exerciseIDs) {
			numExercises = len(exerciseIDs)
		}

		rand.Shuffle(len(exerciseIDs), func(i, j int) {
			exerciseIDs[i], exerciseIDs[j] = exerciseIDs[j], exerciseIDs[i]
		})

		exercisesToLog := exerciseIDs[:numExercises]
		var mostRecentLogDate pgtype.Timestamptz

		// Log random exercises first
		for _, exerciseID := range exercisesToLog {
			exerciseLog := db.ExerciseLog{
				UserID:     userID,
				ExerciseID: exerciseID,
				Reps:       int32(gofakeit.Number(1, 20)),
				Weight:     pgtype.Numeric{Int: big.NewInt(int64(gofakeit.Number(10, 100))), Valid: true},
				LogDate:    pgtype.Timestamptz{Time: randomDateBetween(startDate, endDate), Valid: true},
				CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}

			bodyWeight := pgtype.Numeric{Int: big.NewInt(int64(gofakeit.Number(50, 150))), Valid: true}

			// Insert bodyweight log first
			var bodyweightID int32
			err := dbConn.QueryRow(ctx, `
				INSERT INTO bodyweight_logs (user_id, bodyweight, log_date, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (user_id, log_date) DO UPDATE
				SET bodyweight = EXCLUDED.bodyweight,
					updated_at = EXCLUDED.updated_at
				RETURNING id`, userID, bodyWeight, exerciseLog.LogDate, exerciseLog.CreatedAt, exerciseLog.UpdatedAt).Scan(&bodyweightID)

			if err != nil {
				log.Printf("Error inserting or updating bodyweight log: %v", err)
				continue
			}

			// Insert exercise log with bodyweight_id
			_, err = dbConn.Exec(ctx, `
                INSERT INTO exercise_logs (user_id, exercise_id, reps, weight, bodyweight_id, log_date, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (user_id, exercise_id, log_date) DO UPDATE
				SET reps = EXCLUDED.reps,
					weight = EXCLUDED.weight,
					updated_at = EXCLUDED.updated_at`, userID, exerciseID, exerciseLog.Reps, exerciseLog.Weight, bodyweightID, exerciseLog.LogDate, exerciseLog.CreatedAt, exerciseLog.UpdatedAt)

			if err != nil {
				log.Printf("Error inserting exercise log: %v", err)
			}

			// Update most recent log date
			if exerciseLog.LogDate.Time.After(mostRecentLogDate.Time) {
				mostRecentLogDate = exerciseLog.LogDate
			}
		}

		// Log all exercises with the most recent log date
		bodyWeight := pgtype.Numeric{Int: big.NewInt(int64(gofakeit.Number(50, 150))), Valid: true}

		// Insert bodyweight log with the most recent log date
		var bodyweightID int32
		err := dbConn.QueryRow(ctx, `
			INSERT INTO bodyweight_logs (user_id, bodyweight, log_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, log_date) DO UPDATE
			SET bodyweight = EXCLUDED.bodyweight,
				updated_at = EXCLUDED.updated_at
			RETURNING id`, userID, bodyWeight, mostRecentLogDate, pgtype.Timestamptz{Time: time.Now(), Valid: true}, pgtype.Timestamptz{Time: time.Now(), Valid: true}).Scan(&bodyweightID)

		if err != nil {
			log.Printf("Error inserting or updating bodyweight log: %v", err)
			continue
		}

		// Insert all exercise logs with the same most recent log date
		for _, exerciseID := range exerciseIDs {
			_, err = dbConn.Exec(ctx, `
                INSERT INTO exercise_logs (user_id, exercise_id, reps, weight, bodyweight_id, log_date, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (user_id, exercise_id, log_date) DO UPDATE
				SET reps = EXCLUDED.reps,
					weight = EXCLUDED.weight,
					updated_at = EXCLUDED.updated_at`, userID, exerciseID, int32(gofakeit.Number(1, 20)), pgtype.Numeric{Int: big.NewInt(int64(gofakeit.Number(10, 100))), Valid: true}, bodyweightID, mostRecentLogDate, pgtype.Timestamptz{Time: time.Now(), Valid: true}, pgtype.Timestamptz{Time: time.Now(), Valid: true})

			if err != nil {
				log.Printf("Error inserting exercise log: %v", err)
			}
		}
	}
}

func insertUserTrophies(ctx context.Context, dbConn *pgxpool.Pool, userIDs []int32) {
	// Fetch trophy IDs from the database
	trophyIDs := make([]int32, 0)
	rows, err := dbConn.Query(ctx, `SELECT id FROM trophies`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var trophyID int32
		if err := rows.Scan(&trophyID); err != nil {
			log.Fatal(err)
		}
		trophyIDs = append(trophyIDs, trophyID)
	}

	for _, userID := range userIDs {
		selectedTrophyIDs := shuffleAndSelectTrophies(trophyIDs, 3) // Select up to 3 random trophies
		for i, trophyID := range selectedTrophyIDs {
			displayOrder := pgtype.Int4{Int32: int32(i), Valid: true}
			userTrophy := db.UserTrophy{
				UserID:       userID,
				TrophyID:     trophyID,
				DisplayOrder: displayOrder,
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}

			_, err := dbConn.Exec(ctx, `
                INSERT INTO user_trophies (user_id, trophy_id, display_order, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5)`, userTrophy.UserID, userTrophy.TrophyID, userTrophy.DisplayOrder, userTrophy.CreatedAt, userTrophy.UpdatedAt)
			if err != nil {
				log.Printf("Error inserting user trophy %v: %v", userTrophy, err)
			}
		}
	}
}

func shuffleAndSelectTrophies(trophyIDs []int32, count int) []int32 {
	// Convert to []int for shuffling
	intTrophyIDs := make([]int, len(trophyIDs))
	for i, id := range trophyIDs {
		intTrophyIDs[i] = int(id)
	}

	// Shuffle the slice
	gofakeit.ShuffleInts(intTrophyIDs)

	// Select up to count trophies and convert back to []int32
	selected := make([]int32, 0, count)
	for i := 0; i < count && i < len(intTrophyIDs); i++ {
		selected = append(selected, int32(intTrophyIDs[i]))
	}

	return selected
}

func insertBodyweightLogs(ctx context.Context, dbConn *pgxpool.Pool, userIDs []int32) {
	for _, userID := range userIDs {
		logDate := pgtype.Timestamptz{Time: randomDateBetween(startDate, endDate), Valid: true}
		bodyweightLog := db.BodyweightLog{
			UserID:     userID,
			Bodyweight: pgtype.Numeric{Int: big.NewInt(int64(gofakeit.Number(50, 150))), Valid: true},
			LogDate:    logDate,
			CreatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		_, err := dbConn.Exec(ctx, `
            INSERT INTO bodyweight_logs (user_id, bodyweight, log_date, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5)`, bodyweightLog.UserID, bodyweightLog.Bodyweight, bodyweightLog.LogDate, bodyweightLog.CreatedAt, bodyweightLog.UpdatedAt)
		if err != nil {
			log.Printf("Error inserting bodyweight log %v: %v", bodyweightLog, err)
		}
	}
}
