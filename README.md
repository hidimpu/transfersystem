## Transfersystem API (Go + PostgreSQL)

A simple RESTful API for managing account creation and fund transfers between accounts using Golang, PostgreSQL, and Chi router.

---

## ✅ Features

- Create account
- Get account balance
- Transfer money between accounts
- Basic unit tests

---

## 📦 Requirements

- Go 1.21+
- PostgreSQL
- Git

---

## 📁 Folder Structure

```
transfersystem/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── api/                 # HTTP handlers
│   ├── db/                  # DB setup and schema
│   └── model/               # Structs for accounts & transactions
├── go.mod
├── go.sum
└── .env                    # Database connection string
```

---

## ⚙️ Setup Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/hidimpu/transfersystem.git
cd transfersystem
```

### 2. Set Environment Variables

Create a `.env` file:

```
DB_URL=postgres://postgres:yourpassword@localhost:5432/transfersystem?sslmode=disable
PORT=8080
```

### 3. Initialize Database

Start PostgreSQL and run:

```bash
psql -U postgres -d transfersystem -f internal/db/schema.sql
```

### 4. Run the API

```bash
go run cmd/main.go
```

---

## 🧪 Running Unit Tests

```bash
go test ./...
```

---

## 📥 Sample Data Insert Query

```sql
INSERT INTO accounts (id, balance) VALUES (101, 1000.00), (102, 500.00);
```

---

## 📡 Sample CURL Requests

### Create Account

```bash
curl -X POST http://localhost:8080/accounts \
     -H "Content-Type: application/json" \
     -d '{"account_id": 201, "balance": 500.00}'
```

### Get Account

```bash
curl http://localhost:8080/accounts/201
```

### Create Transaction

```bash
curl -X POST http://localhost:8080/transactions \
     -H "Content-Type: application/json" \
     -d '{"source_account_id": 101, "destination_account_id": 102, "amount": 150.00}'
```

---

## 🧠 Key Points:

- ✅ Clean, modular folder structure (cmd, internal/{api,model,db})
- ✅ Uses decimal for safe financial calculations
- ✅ Uses Chi router for simplicity and performance
- ✅ Follows Go best practices: idiomatic error handling, handler composition
- 🆕 Ready to scale with ease

---
