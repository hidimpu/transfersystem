# Internal Transfers System

## Project Overview
A sample Go HTTP+Postgres API for transferring funds between accounts.

## Local Setup

1. Set up Postgres and create a database named 'transfers'.
2. Export DATABASE_URL="postgres://postgres@localhost:5432/transfers?sslmode=disable"
3. Initialize Go modules: go mod init github.com/hidimpu/transfersystem
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
