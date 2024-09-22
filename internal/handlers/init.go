package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"new-chainsaw/db"
)

var queries *db.Queries

func InitializeQueries(dbPool *pgxpool.Pool) {
	queries = db.New(dbPool)
}
