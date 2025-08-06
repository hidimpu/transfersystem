#!/bin/bash

mkdir -p transfersystem/cmd
mkdir -p transfersystem/internal/{api,service,repository,model,db,config}

cat > transfersystem/cmd/main.go <<'EOF'
package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "transfersystem/internal/config"
    "transfersystem/internal/db"
)

func main() {
    cfg := config.Load()
    database, err := db.NewDB(cfg.DBUrl)
    if err != nil {
        log.Fatal(err)
    }
    defer database.Close()
    r := mux.NewRouter()
    // TODO: Initialize repos, services, handlers, set up routes
    log.Println("Listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
EOF

cat > transfersystem/internal/config/config.go <<'EOF'
package config

import "os"

type Config struct {
    DBUrl string
}

func Load() *Config {
    return &Config{
        DBUrl: os.Getenv("DATABASE_URL"),
    }
}
EOF

cat > transfersystem/internal/db/postgres.go <<'EOF'
package db

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewDB(connStr string) (*sql.DB, error) {
    return sql.Open("postgres", connStr)
}
EOF

cat > transfersystem/internal/model/account.go <<'EOF'
package model

import "github.com/shopspring/decimal"

type Account struct {
    AccountID int64           `json:"account_id"`
    Balance   decimal.Decimal `json:"balance"`
}
EOF

cat > transfersystem/internal/model/transaction.go <<'EOF'
package model

import "github.com/shopspring/decimal"

type Transaction struct {
    SourceAccountID      int64           `json:"source_account_id"`
    DestinationAccountID int64           `json:"destination_account_id"`
    Amount               decimal.Decimal `json:"amount"`
}
EOF

cat > transfersystem/README.md <<'EOF'
# Internal Transfers System

## Project Overview
A sample Go HTTP+Postgres API for transferring funds between accounts.

## Local Setup

1. Set up Postgres and create a database named 'transfers'.
2. Export DATABASE_URL="postgres://postgres@localhost:5432/transfers?sslmode=disable"
3. Initialize Go modules: go mod init github.com/yourusername/transfersystem
4. Install dependencies:
    go get github.com/lib/pq
    go get github.com/gorilla/mux
    go get github.com/shopspring/decimal
5. `go run ./cmd/main.go`

## Assumptions
- Same currency
- No auth
- Simple validation

...
EOF

echo "Project structure and sample files created under ./transfersystem/"
