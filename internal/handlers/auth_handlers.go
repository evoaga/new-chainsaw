package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"log"
	"net/http"
	"time"
	"new-chainsaw/db"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/middleware"
	"new-chainsaw/internal/response"
)

func AuthHandler(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.LogErrorAndRespond(c, http.StatusBadRequest, "No provider found", "You must select a provider", nil)
		return
	}

	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func AuthCallback(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		response.LogErrorAndRespond(c, http.StatusInternalServerError, "Failed to complete user authentication: "+err.Error(), "Failed to complete user authentication", err)
		return
	}

	userID, username, name, isNewUser, err := saveUserToDB(user)
	if err != nil {
		response.LogErrorAndRespond(c, http.StatusInternalServerError, "Failed to save user to database: "+err.Error(), "Failed to save user to database", err)
		return
	}

	fmt.Println("User ID: ", userID)
	fmt.Println("Username: ", username)
	fmt.Println("Name: ", name)

	token, err := middleware.GenerateJWT(userID, username, name, user.Email, user.AvatarURL, isNewUser)
	if err != nil {
		response.LogErrorAndRespond(c, http.StatusInternalServerError, "Failed to generate JWT token: "+err.Error(), "Failed to generate JWT token", err)
		return
	}

	refreshToken, err := generateRefreshToken(userID)
	if err != nil {
		response.LogErrorAndRespond(c, http.StatusInternalServerError, "Failed to generate refresh token: "+err.Error(), "Failed to generate refresh token", err)
		return
	}

	middleware.SetCookies(c, token, refreshToken)
	redirectURL := config.EnvVars["REDIRECT_URL"]
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)

	c.Set("is_new_user", isNewUser)
}

func generateRefreshToken(userID int) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	refreshToken := base64.URLEncoding.EncodeToString(token)

	now := time.Now().UTC()
	expiresAt := now.Add(config.RefreshTokenExpiration)

	err := queries.InsertOrUpdateRefreshToken(context.Background(), db.InsertOrUpdateRefreshTokenParams{
		UserID:    int32(userID),
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	})

	return refreshToken, err
}

func generateUniqueUsername(base string) (string, error) {
	for {
		uniquePart := uuid.New().String()[:8]
		username := base + uniquePart
		_, err := queries.GetUserByUsername(context.Background(), username)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return username, nil
			}
			return "", err
		}
	}
}

func saveUserToDB(user goth.User) (int, string, string, bool, error) {
	ctx := context.Background()

	log.Printf("Attempting to find user by email: %s", user.Email)
	u, err := queries.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return createUser(ctx, user)
		}
		log.Printf("Error querying user by email: %v", err)
		return 0, "", "", false, err
	}

	userID := u.ID
	username := u.Username
	name := u.Name.String
	log.Printf("User found")

	err = upsertUserProvider(ctx, user, userID)
	if err != nil {
		log.Printf("Failed to insert/update user provider information: %v", err)
		return 0, "", "", false, err
	}

	log.Printf("User %s saved/updated successfully", username)
	return int(userID), username, name, false, nil
}

func createUser(ctx context.Context, user goth.User) (int, string, string, bool, error) {
	log.Printf("User not found, creating new user: %s", user.Email)
	baseUsername := user.FirstName
	if baseUsername == "" {
		baseUsername = "user"
	}
	username, err := generateUniqueUsername(baseUsername)
	if err != nil {
		log.Printf("Failed to generate unique username: %v", err)
		return 0, "", "", false, err
	}
	log.Printf("Generated unique username: %s", username)

	userID, err := insertNewUser(ctx, user, username)
	if err != nil {
		return 0, "", "", false, err
	}

	err = insertInitialUserProvider(ctx, user, userID)
	if err != nil {
		return 0, "", "", false, err
	}

	return int(userID), username, "", true, nil
}

func insertNewUser(ctx context.Context, user goth.User, username string) (int32, error) {
	preferredUnits := "metric" // default unit system

	newUser, err := queries.InsertUser(ctx, db.InsertUserParams{
		Username:       username,
		Email:          user.Email,
		AvatarUrl:      pgtype.Text{String: user.AvatarURL, Valid: true},
		PreferredUnits: db.UnitSystem(preferredUnits),
		CreatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		log.Printf("Failed to insert new user into database: %v", err)
		return 0, err
	}
	return newUser.ID, nil
}

func insertInitialUserProvider(ctx context.Context, user goth.User, userID int32) error {
	return queries.InsertInitialUserProvider(ctx, db.InsertInitialUserProviderParams{
		UserID:         userID,
		Provider:       user.Provider,
		ProviderUserID: user.UserID,
		FirstName:      pgtype.Text{String: user.FirstName, Valid: true},
		LastName:       pgtype.Text{String: user.LastName, Valid: true},
		Nickname:       pgtype.Text{String: user.NickName, Valid: true},
		AvatarUrl:      pgtype.Text{String: user.AvatarURL, Valid: true},
		Location:       pgtype.Text{String: user.Location, Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
}

func upsertUserProvider(ctx context.Context, user goth.User, userID int32) error {
	log.Printf("Inserting/updating user provider information for userID: %d", userID)
	return queries.UpsertUserProvider(ctx, db.UpsertUserProviderParams{
		UserID:         userID,
		Provider:       user.Provider,
		ProviderUserID: user.UserID,
		FirstName:      pgtype.Text{String: user.FirstName, Valid: true},
		LastName:       pgtype.Text{String: user.LastName, Valid: true},
		Nickname:       pgtype.Text{String: user.NickName, Valid: true},
		AvatarUrl:      pgtype.Text{String: user.AvatarURL, Valid: true},
		Location:       pgtype.Text{String: user.Location, Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
}
