package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"time"
	"new-chainsaw/db"
	"new-chainsaw/internal/binding"
	"new-chainsaw/internal/middleware"
	"new-chainsaw/internal/response"
	"new-chainsaw/internal/validation"
)

func GetUserProfileByUsernameHandler(c *gin.Context) {
	username := c.Param("username")

	if err := validation.ValidateUsername(username); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, err.Error(), nil, err)
		return
	}

	userProfile, err := queries.GetUserProfileByUsername(context.Background(), username)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.JSONResponse(c, http.StatusNotFound, "User not found", nil, err)
			return
		}
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch user profile", nil, err)
		return
	}

	response.JSONResponse(c, http.StatusOK, "", gin.H{"user": userProfile}, err)
}

func GetUserProfileByIDHandler(c *gin.Context) {
	userID := c.GetInt("userID")

	userProfile, err := queries.GetUserProfileByID(context.Background(), int32(userID))
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.JSONResponse(c, http.StatusNotFound, "User not found", nil, err)
			return
		}
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch user profile", nil, err)
		return
	}

	response.JSONResponse(c, http.StatusOK, "", gin.H{"user": userProfile}, err)
}

func DeleteAccountHandler(c *gin.Context) {
	userID := c.GetInt("userID")

	// Delete refresh tokens associated with the user
	err := queries.DeleteAllRefreshTokensForUser(context.Background(), int32(userID))
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to delete refresh tokens", nil, err)
		return
	}

	// Delete the user
	err = queries.DeleteUser(context.Background(), int32(userID))
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to delete account", nil, err)
		return
	}

	middleware.ClearCookies(c)
	response.JSONResponse(c, http.StatusOK, "Account deleted successfully", nil, err)
}

func CheckUsernameAvailabilityHandler(c *gin.Context) {
	username := c.Query("username")

	if err := validation.ValidateUsername(username); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, err.Error(), nil, err)
		return
	}

	_, err := queries.GetUserByUsername(context.Background(), username)
	if err != nil {
		if err.Error() == "no rows in result set" {
			response.JSONResponse(c, http.StatusOK, "", gin.H{"available": true}, err)
			return
		}
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to check username", nil, err)
		return
	}

	response.JSONResponse(c, http.StatusOK, "", gin.H{"available": false}, err)
}

func UpdateUserHandler(c *gin.Context) {
	var req struct {
		Username       string `json:"username"`
		Name           string `json:"name"`
		AvatarUrl      string `json:"avatar_url"`
		Sex            string `json:"sex"`
		PreferredUnits string `json:"preferred_units"`
	}

	if err := binding.BindJSON(c, &req); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, "Invalid request", nil, err)
		return
	}

	if err := validation.ValidateUsername(req.Username); err != nil {
		response.JSONResponse(c, http.StatusBadRequest, err.Error(), nil, nil)
		return
	}

	userID := c.GetInt("userID")

	var preferredUnits db.UnitSystem
	switch req.PreferredUnits {
	case "metric":
		preferredUnits = db.UnitSystemMetric
	case "imperial":
		preferredUnits = db.UnitSystemImperial
	default:
		// Fallback to metric if invalid or empty
		preferredUnits = db.UnitSystemMetric
	}

	err := queries.UpdateUser(context.Background(), db.UpdateUserParams{
		ID:             int32(userID),
		Username:       req.Username,
		Name:           pgtype.Text{String: req.Name, Valid: req.Name != ""},
		AvatarUrl:      pgtype.Text{String: req.AvatarUrl, Valid: req.AvatarUrl != ""},
		Sex:            pgtype.Text{String: req.Sex, Valid: req.Sex != ""},
		PreferredUnits: preferredUnits,
	})

	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to update user", nil, err)
		return
	}

	// Trigger trophy validation
	err = checkAndUpdateUserTrophies(int32(userID))
	if err != nil {
		fmt.Printf("Failed to update trophies for user %d: %v\n", userID, err)
		return
	}

	fmt.Println("User updated successfully")

	response.JSONResponse(c, http.StatusOK, "User updated successfully", nil, err)
}

func SearchUsers(c *gin.Context) {
	query := c.Query("q")

	// Limit query length to 50 characters
	if len(query) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query too long"})
		return
	}

	// Default values for limit and offset
	const defaultLimit = 10
	const defaultOffset = 0

	limit := defaultLimit
	offset := defaultOffset

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	queryText := pgtype.Text{
		String: query,
		Valid:  true,
	}

	users, err := queries.SearchUsers(ctx, db.SearchUsersParams{
		Column1: queryText,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
