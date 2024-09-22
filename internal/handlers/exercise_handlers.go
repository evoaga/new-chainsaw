package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"math/big"
	"net/http"
	"time"
	"new-chainsaw/db"
	"new-chainsaw/internal/binding"
	"new-chainsaw/internal/conversion"
	"new-chainsaw/internal/response"
)

// Convert sql.NullString to db.NullExerciseType
func toNullExerciseType(ns sql.NullString) db.NullExerciseType {
	return db.NullExerciseType{
		ExerciseType: db.ExerciseType(ns.String),
		Valid:        ns.Valid,
	}
}

func GetLatestExercises(c *gin.Context) {
	userID := c.GetInt("userID")

	exercises, err := queries.GetExercisesWithLatestLogDate(context.Background(), int32(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exercises)
}

type ExerciseRequest struct {
	ExerciseID       int32    `json:"exercise_id"`
	Reps             int32    `json:"reps"`
	Weight           float64  `json:"weight"`
	Unit             string   `json:"unit"`
	BodyWeight       float64  `json:"body_weight"`
	LogDate          string   `json:"log_date"`
	AdditionalWeight *float64 `json:"additional_weight"`
	ExerciseType     string   `json:"exercise_type"`
}

type ExerciseResponse struct {
	ExerciseID       int32   `json:"exercise_id"`
	Reps             int32   `json:"reps"`
	Weight           float64 `json:"weight"`
	Unit             string  `json:"unit"`
	LogDate          string  `json:"log_date"`
	AdditionalWeight float64 `json:"additional_weight"`
	ExerciseType     string  `json:"exercise_type"`
}

func LogExerciseHandler(c *gin.Context) {
	var reqs []ExerciseRequest
	if err := binding.BindJSON(c, &reqs); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}

	userID := c.GetInt("userID")
	var duplicates []ExerciseResponse
	var logged []ExerciseResponse

	for _, req := range reqs {
		log.Printf("Received request payload: %+v\n", req)
		if err := processExerciseLog(userID, req, &logged, &duplicates); err != nil {
			response.JSONResponse(c, http.StatusInternalServerError, err.Error(), nil, err)
			return
		}
	}

	// Trigger trophy validation
	err := checkAndUpdateUserTrophies(int32(userID))
	if err != nil {
		fmt.Printf("Failed to update trophies for user %d: %v\n", userID, err)
		return
	}

	response.JSONResponse(c, http.StatusOK, "Exercises and body weight logged successfully", gin.H{"logged": logged, "duplicates": duplicates}, nil)
}

func processExerciseLog(userID int, req ExerciseRequest, logged *[]ExerciseResponse, duplicates *[]ExerciseResponse) error {
	logDate, err := parseLogDate(req.LogDate)
	if err != nil {
		return err
	}

	bodyweightID, err := getOrCreateBodyweightID(userID, req.BodyWeight, req.Unit, logDate)
	if err != nil {
		return err
	}

	if err := logOrUpdateExercise(userID, req, logDate, bodyweightID, logged, duplicates); err != nil {
		return err
	}

	return nil
}

func parseLogDate(dateStr string) (time.Time, error) {
	logDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Error parsing date: %v\n", err)
		return time.Time{}, errors.New("invalid date format")
	}
	return logDate, nil
}

type ExerciseLogRequest struct {
	userID       int
	req          ExerciseRequest
	logDate      time.Time
	bodyweightID int32
}

type ExerciseLogData struct {
	weight           pgtype.Numeric
	additionalWeight pgtype.Numeric
	exerciseType     db.NullExerciseType
}

func logOrUpdateExercise(userID int, req ExerciseRequest, logDate time.Time, bodyweightID int32, logged *[]ExerciseResponse, duplicates *[]ExerciseResponse) error {
	data := ExerciseLogRequest{
		userID:       userID,
		req:          req,
		logDate:      logDate,
		bodyweightID: bodyweightID,
	}

	weight := convertToPgNumeric(calculateWeight(req.Weight, req.Unit))
	additionalWeight := convertAdditionalWeight(req.AdditionalWeight)
	exerciseType := toNullExerciseType(sql.NullString{String: req.ExerciseType, Valid: req.ExerciseType != ""})

	logData := ExerciseLogData{
		weight:           weight,
		additionalWeight: additionalWeight,
		exerciseType:     exerciseType,
	}

	// Try to insert the exercise log
	err := queries.LogExercise(context.Background(), db.LogExerciseParams{
		UserID:           int32(data.userID),
		ExerciseID:       data.req.ExerciseID,
		Reps:             data.req.Reps,
		Weight:           logData.weight,
		AdditionalWeight: logData.additionalWeight,
		ExerciseType:     logData.exerciseType,
		BodyweightID:     data.bodyweightID,
		LogDate:          pgtype.Timestamptz{Time: data.logDate, Valid: true},
	})

	if err != nil {
		return handleExerciseLogError(err, data, logData, logged, duplicates)
	}

	appendToLogged(logged, data.req, logData.additionalWeight)
	return nil
}

func handleExerciseLogError(err error, data ExerciseLogRequest, logData ExerciseLogData, logged *[]ExerciseResponse, duplicates *[]ExerciseResponse) error {
	if isUniqueViolation(err) {
		appendToDuplicates(duplicates, data.req)
		updateErr := queries.UpdateExerciseLog(context.Background(), db.UpdateExerciseLogParams{
			UserID:           int32(data.userID),
			ExerciseID:       data.req.ExerciseID,
			Reps:             data.req.Reps,
			Weight:           logData.weight,
			AdditionalWeight: logData.additionalWeight,
			ExerciseType:     logData.exerciseType,
			BodyweightID:     data.bodyweightID,
			LogDate:          pgtype.Timestamptz{Time: data.logDate, Valid: true},
		})
		if updateErr != nil {
			return errors.New("failed to update exercise log")
		}

		appendToLogged(logged, data.req, logData.additionalWeight)
		return nil
	}
	log.Printf("Error logging exercise: %v\n", err)
	return errors.New("failed to log exercise")
}

func calculateWeight(weight float64, unit string) float64 {
	if unit == "imperial" {
		return conversion.LbsToKg(weight)
	}
	return weight
}

func convertAdditionalWeight(additionalWeight *float64) pgtype.Numeric {
	if additionalWeight != nil {
		return convertToPgNumeric(*additionalWeight)
	}
	return pgtype.Numeric{Valid: false}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func appendToDuplicates(duplicates *[]ExerciseResponse, req ExerciseRequest) {
	var additionalWeightVal float64
	if req.AdditionalWeight != nil {
		additionalWeightVal = *req.AdditionalWeight
	} else {
		additionalWeightVal = 0.0
	}

	*duplicates = append(*duplicates, ExerciseResponse{
		ExerciseID:       req.ExerciseID,
		Reps:             req.Reps,
		Weight:           req.Weight,
		Unit:             req.Unit,
		LogDate:          req.LogDate,
		AdditionalWeight: additionalWeightVal,
		ExerciseType:     req.ExerciseType,
	})
}

func appendToLogged(logged *[]ExerciseResponse, req ExerciseRequest, additionalWeight pgtype.Numeric) {
	var additionalWeightVal float64
	if additionalWeight.Valid {
		val, _ := additionalWeight.Int.Float64()
		additionalWeightVal = val
	} else {
		additionalWeightVal = 0.0
	}

	*logged = append(*logged, ExerciseResponse{
		ExerciseID:       req.ExerciseID,
		Reps:             req.Reps,
		Weight:           req.Weight,
		Unit:             req.Unit,
		LogDate:          req.LogDate,
		AdditionalWeight: additionalWeightVal,
		ExerciseType:     req.ExerciseType,
	})
}

func getOrCreateBodyweightID(userID int, bodyWeight float64, unit string, logDate time.Time) (int32, error) {
	bodyWeightKg := bodyWeight
	if unit == "imperial" {
		bodyWeightKg = conversion.LbsToKg(bodyWeight)
	}

	weight := convertToPgNumeric(bodyWeightKg)
	bodyweightID, err := queries.LogBodyWeight(context.Background(), db.LogBodyWeightParams{
		UserID:     int32(userID),
		Bodyweight: weight,
		LogDate:    pgtype.Timestamptz{Time: logDate, Valid: true},
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Printf("Conflict logging body weight, attempting to update: %v\n", err)
			bodyweightID, err = fetchBodyweightID(userID, logDate)
			if err != nil {
				return 0, err
			}
			// Update the existing bodyweight record
			if updateErr := updateBodyWeight(userID, weight, logDate); updateErr != nil {
				return 0, updateErr
			}
			return bodyweightID, nil
		}
		return 0, handleBodyWeightError(err, userID, weight, logDate)
	}
	return bodyweightID, nil
}

func fetchBodyweightID(userID int, logDate time.Time) (int32, error) {
	id, err := queries.GetBodyweightLogByUserIDAndDate(context.Background(), db.GetBodyweightLogByUserIDAndDateParams{
		UserID:  int32(userID),
		LogDate: pgtype.Timestamptz{Time: logDate, Valid: true},
	})

	if err != nil {
		log.Printf("Error fetching existing body weight ID: %v\n", err)
		return 0, errors.New("failed to fetch existing body weight ID")
	}
	return id, nil
}

func handleBodyWeightError(err error, userID int, weight pgtype.Numeric, logDate time.Time) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		log.Printf("Conflict logging body weight, attempting to update: %v\n", err)
		return updateBodyWeight(userID, weight, logDate)
	}
	log.Printf("Error logging body weight: %v\n", err)
	return errors.New("failed to log body weight")
}

func updateBodyWeight(userID int, weight pgtype.Numeric, logDate time.Time) error {
	err := queries.UpdateBodyWeight(context.Background(), db.UpdateBodyWeightParams{
		UserID:     int32(userID),
		Bodyweight: weight,
		LogDate:    pgtype.Timestamptz{Time: logDate, Valid: true},
	})

	if err != nil {
		log.Printf("Error updating body weight: %v\n", err)
		return errors.New("failed to log body weight")
	}
	return nil
}

func convertToPgNumeric(value float64) pgtype.Numeric {
	num := pgtype.Numeric{
		Int:   new(big.Int),
		Exp:   0,
		NaN:   false,
		Valid: true,
	}
	num.Int.SetString(big.NewFloat(value).Text('f', -1), 10)
	return num
}
