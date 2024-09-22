package auth

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"new-chainsaw/internal/config"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

func ConfigureSessionStore(sessionSecret string) {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store
}

func InitOAuth() {
	backendURL := config.EnvVars["BACKEND_URL"]

	googleCallbackURL := fmt.Sprintf("%s/auth/google/callback", backendURL)
	githubCallbackURL := fmt.Sprintf("%s/auth/github/callback", backendURL)
	appleCallbackURL := fmt.Sprintf("%s/auth/apple/callback", backendURL)

	goth.UseProviders(
		google.New(config.EnvVars["GOOGLE_KEY"], config.EnvVars["GOOGLE_SECRET"], googleCallbackURL),
		github.New(config.EnvVars["GITHUB_KEY"], config.EnvVars["GITHUB_SECRET"], githubCallbackURL),
		apple.New(config.EnvVars["APPLE_KEY"], config.EnvVars["APPLE_SECRET"], appleCallbackURL, nil, apple.ScopeName, apple.ScopeEmail),
	)
}
