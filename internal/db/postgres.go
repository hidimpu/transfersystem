package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	dbURL := os.Getenv("DB_URL")
	return sql.Open("postgres", dbURL)
}
