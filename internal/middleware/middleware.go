package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"net/http"
	"time"
	"new-chainsaw/db"
	"new-chainsaw/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Claims struct {
	UserID    int    `json:"sub"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	IsNewUser bool   `json:"is_new_user"`
	jwt.RegisteredClaims
}

func JWTMiddleware(dbPool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(config.JwtCookieName)
		if err != nil || tokenString == "" {
			log.Println("No JWT token provided, checking refresh token")
			handleRefreshToken(c, dbPool)
			return
		}

		jwtSecret := config.EnvVars["JWT_SECRET"]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Error parsing JWT token: %v", err)
			handleRefreshToken(c, dbPool)
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("avatar_url", claims.AvatarURL)
		c.Set("name", claims.Name)
		c.Set("is_new_user", claims.IsNewUser)

		c.Next()
	}
}

func handleRefreshToken(c *gin.Context, dbPool *pgxpool.Pool) {
	refreshToken, err := c.Cookie(config.RefreshTokenCookie)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
		return
	}

	queries := db.New(dbPool)
	user, err := queries.ValidateRefreshToken(context.Background(), refreshToken)
	if err != nil {
		log.Printf("Failed to validate refresh token: %v, Token: %s", err, refreshToken)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		c.Abort()
		return
	}

	// Assume isNewUser is false for refresh tokens since we don't store it in the database
	token, err := GenerateJWT(int(user.ID), user.Username, user.Name.String, user.Email, user.AvatarUrl.String, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token"})
		c.Abort()
		return
	}

	newRefreshToken, err := generateRefreshToken(int(user.ID), dbPool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		c.Abort()
		return
	}

	SetCookies(c, token, newRefreshToken)
	c.Set("userID", int(user.ID))
	c.Set("username", user.Username)
	c.Set("email", user.Email)
	c.Set("avatar_url", user.AvatarUrl.String)
	c.Set("name", user.Name.String)

	c.Next()
}

func GenerateJWT(userID int, username, name, email, avatarURL string, isNewUser bool) (string, error) {
	expirationTime := time.Now().Add(config.JwtExpiration)
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		AvatarURL: avatarURL,
		Name:      name,
		IsNewUser: isNewUser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := config.EnvVars["JWT_SECRET"]

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateRefreshToken(userID int, dbPool *pgxpool.Pool) (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	refreshToken := base64.URLEncoding.EncodeToString(token)

	expiresAt := time.Now().Add(config.RefreshTokenExpiration)

	queries := db.New(dbPool)
	params := db.InsertOrUpdateRefreshTokenParams{
		UserID:    int32(userID),
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	}
	err := queries.InsertOrUpdateRefreshToken(context.Background(), params)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func SetCookies(c *gin.Context, jwtToken, refreshToken string) {
	isProduction := config.EnvVars["ENV"] == "production"
	domain := config.EnvVars["FRONTEND_URL"]

	c.SetCookie(config.JwtCookieName, jwtToken, int(config.JwtExpiration.Seconds()), "/", domain, isProduction, true)
	c.SetCookie(config.RefreshTokenCookie, refreshToken, int(config.RefreshTokenExpiration.Seconds()), "/", domain, isProduction, true)
}

func ClearCookies(c *gin.Context) {
	isProduction := config.EnvVars["ENV"] == "production"
	domain := config.EnvVars["FRONTEND_URL"]

	c.SetCookie(config.JwtCookieName, "", -1, "/", domain, isProduction, true)
	c.SetCookie(config.RefreshTokenCookie, "", -1, "/", domain, isProduction, true)
}
