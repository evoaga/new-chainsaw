package config

import (
	"log"
	"os"
	"strings"
	"time"
)

const (
	JwtCookieName          = "jwt"
	RefreshTokenCookie     = "refresh_token"
	JwtExpiration          = 1 * time.Hour
	RefreshTokenExpiration = 7 * 24 * time.Hour
)

var EnvVars = make(map[string]string)

func LoadConfig() {
	criticalVars := []string{
		"DB_HOST", "DB_PORT", "DB_DATABASE", "DB_USERNAME", "DB_PASSWORD",
		"JWT_SECRET", "REFRESH_SECRET",
		"FRONTEND_URL", "BACKEND_URL",
	}
	var missingVars []string

	for _, envVar := range os.Environ() {
		pair := strings.SplitN(envVar, "=", 2)
		if len(pair) == 2 {
			EnvVars[pair[0]] = pair[1]
		}
	}

	for _, key := range criticalVars {
		if value, exists := EnvVars[key]; !exists || value == "" {
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) > 0 {
		log.Fatalf("Critical environment variables missing: %v", missingVars)
	}
}
