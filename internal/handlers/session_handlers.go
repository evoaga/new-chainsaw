package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"new-chainsaw/db"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/middleware"
	"new-chainsaw/internal/response"
)

func SessionHandler(c *gin.Context) {
	username := c.GetString("username")
	email := c.GetString("email")
	avatarURL := c.GetString("avatar_url")
	name := c.GetString("name")
	isNewUser := c.GetBool("is_new_user")

	sessionData := map[string]interface{}{
		"username":    username,
		"email":       email,
		"avatar_url":  avatarURL,
		"name":        name,
		"is_new_user": isNewUser,
	}

	userID := c.GetInt("userID")
	prefs, err := queries.GetUserPreferences(c, int32(userID))
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch preferences", nil, err)
		return
	}

	preferencesData := map[string]interface{}{
		"sex":             prefs.Sex,
		"preferred_units": prefs.PreferredUnits,
	}

	response.JSONResponse(c, http.StatusOK, "", gin.H{
		"session":     sessionData,
		"preferences": preferencesData,
	}, nil)
}

func SignOutHandler(c *gin.Context) {
	userID := c.GetInt("userID")

	refreshToken, err := c.Cookie(config.RefreshTokenCookie)
	if err != nil {
		response.JSONResponse(c, http.StatusBadRequest, "No refresh token provided", nil, err)
		return
	}

	err = queries.DeleteRefreshToken(context.Background(), db.DeleteRefreshTokenParams{
		UserID: int32(userID),
		Token:  refreshToken,
	})
	if err != nil {
		response.JSONResponse(c, http.StatusInternalServerError, "Failed to sign out", nil, err)
		return
	}

	middleware.ClearCookies(c)
	response.JSONResponse(c, http.StatusOK, "Signed out successfully", nil, err)
}
