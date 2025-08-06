package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewDB(connStr string) (*sql.DB, error) {
	return sql.Open("postgres", connStr)
}
