package server

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"net/http"
	"os"
	"strconv"
	"time"
	"new-chainsaw/internal/auth"
	"new-chainsaw/internal/config"
	"new-chainsaw/internal/handlers"

	"new-chainsaw/internal/database"
)

type Server struct {
	port int

	db database.Service

	dbPool *pgxpool.Pool
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	dbService := database.New()
	dbPool := dbService.GetDB()

	NewServer := &Server{
		port: port,

		db: dbService,

		dbPool: dbPool,
	}

	// Initialize the handlers with db pool
	handlers.InitializeQueries(dbPool)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Configure session store
	sessionSecret := config.EnvVars["SESSION_SECRET"]
	auth.ConfigureSessionStore(sessionSecret)

	// Initialize OAuth
	auth.InitOAuth()

	return server
}
