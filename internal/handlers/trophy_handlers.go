package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"new-chainsaw/db"
	"new-chainsaw/internal/binding"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/httpclient"
	"new-chainsaw/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type TrophyRequest struct {
	TrophyID     int32 `json:"trophy_id"`
	DisplayOrder int32 `json:"display_order"`
}

type ValidateTrophiesRequest struct {
	Trophies []TrophyRequest `json:"trophies"`
}

func ValidateAndSaveTrophiesHandler(c *gin.Context) {
	var req ValidateTrophiesRequest

	if err := binding.BindJSON(c, &req); err != nil {
		return
	}

	userID := c.GetInt("userID")

	fmt.Println(req)

	// Ensure there are at most 3 trophies
	if len(req.Trophies) > 3 {
		response.JSONResponse(c, http.StatusBadRequest, "Cannot display more than 3 trophies", nil, nil)
		return
	}

	// Retrieve all user details in one go
	userDetails, err := queries.GetUserDetails(context.Background(), int32(userID))
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch user details", nil, err)
		return
	}

	// Convert ExerciseLogs interface{} to JSON for further processing
	exerciseLogsJSON, err := json.Marshal(userDetails.ExerciseLogs)
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to process exercise logs", nil, err)
		return
	}

	var exerciseLogs []map[string]interface{}
	if err := json.Unmarshal(exerciseLogsJSON, &exerciseLogs); err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to process exercise logs", nil, err)
		return
	}

	// Call validateTrophies to get evaluated trophies
	evaluatedTrophies, err := validateTrophies(req, userDetails.Sex, userDetails.LatestBodyWeight, userDetails.PreferredUnits, exerciseLogs)
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	// Filter unlocked trophies
	var unlockedTrophies []TrophyRequest
	for _, trophy := range evaluatedTrophies {
		if unlocked, ok := trophy["unlocked"].(bool); ok && unlocked {
			if trophyID, ok := trophy["trophy_id"].(float64); ok { // trophy_id is likely a float64 due to JSON unmarshalling
				for _, reqTrophy := range req.Trophies {
					if int32(trophyID) == reqTrophy.TrophyID {
						unlockedTrophies = append(unlockedTrophies, reqTrophy)
						break
					}
				}
			}
		}
	}

	// Save unlocked trophies in the database
	for _, trophy := range unlockedTrophies {
		// Check if there is an existing trophy with the same display_order
		existingTrophy, err := queries.GetTrophyByDisplayOrder(context.Background(), db.GetTrophyByDisplayOrderParams{
			UserID:       int32(userID),
			DisplayOrder: pgtype.Int4{Int32: trophy.DisplayOrder, Valid: true},
		})
		if err == nil && existingTrophy.ID != 0 {
			// If an existing trophy is found, delete it
			err = queries.DeleteUserTrophy(context.Background(), db.DeleteUserTrophyParams{
				UserID:   int32(userID),
				TrophyID: existingTrophy.ID,
			})
			if err != nil {
				fmt.Printf("Failed to delete old trophy with ID %d: %v\n", existingTrophy.ID, err)
				response.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete old trophy with ID %d", existingTrophy.ID), nil, err)
				return
			}
		}

		// Insert the new trophy
		err = queries.InsertUserTrophy(context.Background(), db.InsertUserTrophyParams{
			UserID:       int32(userID),
			TrophyID:     trophy.TrophyID,
			DisplayOrder: pgtype.Int4{Int32: trophy.DisplayOrder, Valid: true},
		})
		if err != nil {
			fmt.Printf("Failed to save trophy with ID %d: %v\n", trophy.TrophyID, err)
			response.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to save trophy with ID %d", trophy.TrophyID), nil, err)
			return
		}
	}

	response.JSONResponse(c, http.StatusOK, "Trophies saved successfully", nil, err)
}

func checkAndUpdateUserTrophies(userID int32) error {
	start := time.Now()

	// Retrieve all user details in one go
	userDetails, err := queries.GetUserDetails(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user details: %w", err)
	}

	// Convert ExerciseLogs interface{} to JSON for further processing
	exerciseLogsJSON, err := json.Marshal(userDetails.ExerciseLogs)
	if err != nil {
		return fmt.Errorf("failed to process exercise logs: %w", err)
	}

	var exerciseLogs []map[string]interface{}
	if err := json.Unmarshal(exerciseLogsJSON, &exerciseLogs); err != nil {
		return fmt.Errorf("failed to process exercise logs: %w", err)
	}

	// Retrieve user's trophies
	userTrophies, err := queries.GetUserTrophies(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("failed to fetch user trophies: %w", err)
	}

	trophyRequests := make([]TrophyRequest, len(userTrophies))
	for i, trophy := range userTrophies {
		if !trophy.DisplayOrder.Valid {
			return fmt.Errorf("invalid display order for trophy ID %d", trophy.ID)
		}
		trophyRequests[i] = TrophyRequest{
			TrophyID:     trophy.ID,
			DisplayOrder: trophy.DisplayOrder.Int32,
		}
	}

	validateReq := ValidateTrophiesRequest{
		Trophies: trophyRequests,
	}

	// Call validateTrophies to get evaluated trophies
	evaluatedTrophies, err := validateTrophies(validateReq, userDetails.Sex, userDetails.LatestBodyWeight, userDetails.PreferredUnits, exerciseLogs)
	if err != nil {
		return fmt.Errorf("failed to validate trophies: %w", err)
	}

	// Determine which trophies to keep and which to remove
	unlockedTrophies := make(map[int32]TrophyRequest)
	for _, trophy := range evaluatedTrophies {
		if unlocked, ok := trophy["unlocked"].(bool); ok && unlocked {
			if trophyID, ok := trophy["trophy_id"].(float64); ok {
				for _, reqTrophy := range validateReq.Trophies {
					if int32(trophyID) == reqTrophy.TrophyID {
						unlockedTrophies[reqTrophy.TrophyID] = reqTrophy
						break
					}
				}
			}
		}
	}

	// Delete trophies that are no longer unlocked
	for _, trophy := range userTrophies {
		if !trophy.DisplayOrder.Valid {
			continue
		}
		if _, found := unlockedTrophies[trophy.ID]; !found {
			err = queries.DeleteUserTrophy(context.Background(), db.DeleteUserTrophyParams{
				UserID:   userID,
				TrophyID: trophy.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete trophy with ID %d: %w", trophy.ID, err)
			}
		}
	}

	// Save unlocked trophies in the database
	for _, trophy := range unlockedTrophies {
		// Check if there is an existing trophy with the same display_order
		existingTrophy, err := queries.GetTrophyByDisplayOrder(context.Background(), db.GetTrophyByDisplayOrderParams{
			UserID:       userID,
			DisplayOrder: pgtype.Int4{Int32: trophy.DisplayOrder, Valid: true},
		})
		if err == nil && existingTrophy.ID != 0 {
			// If an existing trophy is found, delete it
			err = queries.DeleteUserTrophy(context.Background(), db.DeleteUserTrophyParams{
				UserID:   userID,
				TrophyID: existingTrophy.ID,
			})
			if err != nil {
				return fmt.Errorf("failed to delete old trophy with ID %d: %w", existingTrophy.ID, err)
			}
		}

		// Insert the new trophy
		err = queries.InsertUserTrophy(context.Background(), db.InsertUserTrophyParams{
			UserID:       userID,
			TrophyID:     trophy.TrophyID,
			DisplayOrder: pgtype.Int4{Int32: trophy.DisplayOrder, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to save trophy with ID %d: %w", trophy.TrophyID, err)
		}
	}

	duration := time.Since(start)
	fmt.Printf("checkAndUpdateUserTrophies took %v to complete for user %d\n", duration, userID)

	return nil
}

func validateTrophies(req ValidateTrophiesRequest, userSex pgtype.Text, latestBodyWeight pgtype.Numeric, unit db.UnitSystem, exerciseLogs []map[string]interface{}) ([]map[string]interface{}, error) {
	payload := map[string]interface{}{
		"user": map[string]interface{}{
			"sex":        userSex.String,
			"bodyWeight": latestBodyWeight,
			"units":      unit,
		},
		"exerciseLogs": exerciseLogs,
		"trophies":     req.Trophies,
	}

	frontendFullURL := config.EnvVars["FRONTEND_FULL_URL"]
	nextJSURL := frontendFullURL + "/api/validate-trophy"

	var responseBody map[string]interface{}
	if err := httpclient.PostJSON(nextJSURL, payload, &responseBody); err != nil {
		return nil, fmt.Errorf("user not authorized to display these trophies")
	}

	evaluatedTrophies, ok := responseBody["trophies"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	trophies := make([]map[string]interface{}, len(evaluatedTrophies))
	for i, trophy := range evaluatedTrophies {
		trophies[i] = trophy.(map[string]interface{})
	}

	return trophies, nil
}

type Trophy struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	DisplayOrder int32  `json:"display_order"`
}

func GetTrophiesHandler(c *gin.Context) {
	userID := c.GetInt("userID")

	userTrophies, err := queries.GetUserTrophies(context.Background(), int32(userID))
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch user trophies", nil, err)
		return
	}

	// Map database trophies to the response structure
	var trophies []Trophy
	for _, t := range userTrophies {
		displayOrder := int32(0)
		if t.DisplayOrder.Valid {
			displayOrder = t.DisplayOrder.Int32
		}

		trophies = append(trophies, Trophy{
			ID:           t.ID,
			Name:         t.Name,
			DisplayOrder: displayOrder,
		})
	}

	response.JSONResponse(c, http.StatusOK, "", gin.H{"trophies": trophies}, err)
}

func DeleteTrophy(c *gin.Context) {
	userID := c.GetInt("userID")

	displayOrder, err := strconv.Atoi(c.Param("display_order"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid display order"})
		return
	}

	// Manually set the fields of pgtype.Int4
	displayOrderPgType := pgtype.Int4{
		Int32: int32(displayOrder),
		Valid: true,
	}

	err = queries.DeleteUserTrophyByOrder(context.Background(), db.DeleteUserTrophyByOrderParams{
		UserID:       int32(userID),
		DisplayOrder: displayOrderPgType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("deleted")

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
