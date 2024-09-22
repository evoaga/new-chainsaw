package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// GetDB returns the underlying db pool.
	GetDB() *pgxpool.Pool
}

type service struct {
	db *pgxpool.Pool
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, database)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Printf(fmt.Sprintf("db down: %v", err)) // Log the error
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stat()
	stats["total_connections"] = strconv.Itoa(int(dbStats.TotalConns()))
	stats["idle_connections"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["acquired_connections"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["acquire_count"] = strconv.FormatInt(dbStats.AcquireCount(), 10)
	stats["acquire_duration"] = dbStats.AcquireDuration().String()
	stats["canceled_acquire_count"] = strconv.FormatInt(dbStats.CanceledAcquireCount(), 10)
	stats["constructing_connections"] = strconv.Itoa(int(dbStats.ConstructingConns()))

	// Evaluate stats to provide a health message
	if dbStats.TotalConns() > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.AcquireCount() > 1000 {
		stats["message"] = "The database has a high number of acquire events, indicating potential bottlenecks."
	}

	if dbStats.CanceledAcquireCount() > int64(dbStats.TotalConns())/2 {
		stats["message"] = "Many acquire requests are being canceled, consider revising the connection pool settings."
	}

	if dbStats.ConstructingConns() > dbStats.TotalConns()/2 {
		stats["message"] = "Many connections are being constructed, consider increasing pool size or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	s.db.Close()
	return nil
}

// GetDB returns the underlying db pool.
func (s *service) GetDB() *pgxpool.Pool {
	return s.db
}
