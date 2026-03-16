# Transfersystem API (Go + PostgreSQL)

An internal transfers service that exposes HTTP endpoints for:

- Creating accounts
- Querying account balances
- Submitting internal transfers between accounts

All balances and transfers are persisted in PostgreSQL with strong concurrency
and data integrity guarantees.

---

## 1. Overview & Architecture

### 1.1 High-level design

The application follows a classic layered / hexagonal style architecture:

- **HTTP layer (`internal/api`)**
  - Implements the exercise endpoints:
    - `POST /accounts` – create an account
    - `GET  /accounts/{account_id}` – query account balance
    - `POST /transactions` – submit a transfer
  - Validates and deserialises JSON requests.
  - Maps domain/service errors to HTTP status codes.

- **Service layer (`internal/service`)**
  - Contains business rules and invariants:
    - Account validation (ID > 0, non-negative balances).
    - Transfer validation (positive amount, different accounts, accounts must exist).
    - Concurrency correctness and transaction boundaries.
  - Uses typed domain errors (`AccountError`, `TransferError`) that encode HTTP
    status codes.

- **Repository layer (`internal/repository`)**
  - Encapsulates all SQL and DB access for accounts and transactions.
  - Provides account CRUD and transaction logging.
  - Exposes small, focused methods such as `GetByID`, `Exists`, and
    `UpdateBalanceTx` (which uses `SELECT ... FOR UPDATE`).

- **Database layer (`internal/db`)**
  - Responsible for connecting to PostgreSQL via `DB_URL`.
  - Performs a startup `Ping` so misconfiguration fails fast.

- **Utilities (`internal/utils`)**
  - Simple structured logging helper used across services and handlers.

Entry point: `cmd/main.go` wires these layers together using the Chi router.

### 1.2 Concurrency model (transfers)

Transfers are processed entirely inside a **single SQL transaction** using the
`SERIALIZABLE` isolation level and row-level locks:

1. Validate request (IDs, amount, existence of accounts).
2. Begin DB transaction with `sql.LevelSerializable`.
3. `SELECT ... FOR UPDATE` on both accounts via `UpdateBalanceTx`.
4. Compute new balances with `shopspring/decimal` and ensure they are not
   negative (no overdrafts).
5. `UPDATE` both balances.
6. Insert a row into `transactions` table.
7. Commit the transaction.

If any step fails, the transaction is rolled back and an appropriate typed
`TransferError` is returned and mapped to an HTTP status code (400 / 404 / 422
/ 500).

---

## 2. Requirements

You will need:

- Go **1.21+**
- PostgreSQL **13+**
- Git
- `curl` (for manual testing)
- `psql` (PostgreSQL client) is recommended but optional


---

## 3. Folder Structure

```text
transfersystem/
├── cmd/
│   └── main.go                  # Entry point (wires router, services, repos)
├── internal/
│   ├── api/                     # HTTP handlers (accounts, transactions)
│   ├── config/                  # (Reserved for config helpers)
│   ├── db/                      # DB connection + schema
│   ├── model/                   # Domain models + error types
│   ├── repository/              # Account & transaction repositories
│   ├── service/                 # Business logic services
│   └── utils/                   # Logger and shared utilities
├── internal/db/schema.sql       # Database schema
├── README.md                    # (this file)
├── go.mod
└── go.sum
```

---

## 4. Setup & Installation

The steps below assume a Unix-like environment (macOS/Linux). For Windows,
commands are similar but environment-variable syntax differs slightly.

### 4.1 Clone the repository

```bash
git clone https://github.com/hidimpu/tsystem.git
cd transfersystem
```

The directory should contain `cmd/`, `internal/`, `README.md`.

### 4.2 Start PostgreSQL

Ensure a PostgreSQL instance is running and reachable via the connection string you plan to use.

For example, on macOS with Homebrew:

```bash
brew install postgresql@16
brew services start postgresql@16
```

On Linux distributions you can use the native package manager (e.g. `sudo apt install postgresql`).

Once the server is up, create the exercise database (adjust the command if you use a different superuser):

```bash
createdb transfersystem
```

If needed, create or ensure a `postgres` superuser exists so the connection string in later steps works unchanged.

### 4.3 Apply the schema

From the project root:

```bash
psql "postgres://postgres:postgres@localhost:5432/transfersystem?sslmode=disable" \
  -f internal/db/schema.sql
```

This creates the required tables:

```sql
-- accounts table
CREATE TABLE IF NOT EXISTS accounts (
    account_id BIGINT PRIMARY KEY,
    balance DECIMAL(20,5) NOT NULL DEFAULT 0.00000
);

-- transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount DECIMAL(20,5) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_account_id) REFERENCES accounts(account_id),
    FOREIGN KEY (destination_account_id) REFERENCES accounts(account_id)
);
```

You can confirm this worked:

```bash
psql "postgres://postgres:postgres@localhost:5432/transfersystem?sslmode=disable" -c "\\dt"
```

You should see `accounts` and `transactions` tables listed.

### 4.4 Configure environment

The service reads configuration from environment variables (or a `.env` file):

- `DB_URL` – PostgreSQL connection string
- `PORT`   – HTTP port to listen on (defaults to 8080 if unset)

Create a `.env` file in the root if you prefer not to export variables manually:

```bash
cat > .env <<'EOF'
DB_URL=postgres://postgres:postgres@localhost:5432/transfersystem?sslmode=disable
PORT=8080
EOF
```

`cmd/main.go` uses `godotenv` to load `.env` automatically on startup.

Alternatively, export the variables in your shell:

```bash
export DB_URL="postgres://postgres:postgres@localhost:5432/transfersystem?sslmode=disable"
export PORT=8080
```

`internal/db/postgres.go` validates that `DB_URL` is set and issues an initial `Ping`. If the database is unreachable the process exits with a clear error message.

### 4.5 Run unit tests

From the project root:

```bash
go test ./...
```

All unit tests currently pass:

```text
?    github.com/hidimpu/transfersystem/cmd            [no test files]
?    github.com/hidimpu/transfersystem/internal/api   [no test files]
?    github.com/hidimpu/transfersystem/internal/config [no test files]
?    github.com/hidimpu/transfersystem/internal/db    [no test files]
ok   github.com/hidimpu/transfersystem/internal/model  2.57s
?    github.com/hidimpu/transfersystem/internal/repository [no test files]
ok   github.com/hidimpu/transfersystem/internal/service 1.28s
?    github.com/hidimpu/transfersystem/internal/utils  [no test files]
```

### 4.6 Run the HTTP API

From the project root:

```bash
go run ./cmd/main.go
```

Expected startup logs:

```text
DB Connection Established!
Server started on port: 8080
Database locks: FOR UPDATE with Serializable isolation
Architecture: MVC with proper separation of concerns
Concurrency: Row-level locking with atomic transactions
```

If the server exits immediately with an error, check that:

- `DB_URL` is set correctly.
- Postgres is running and accessible.
- The chosen `PORT` is not already in use.

---

## 5. HTTP API

Base URL (default): `http://localhost:8080`

### 5.1 Create account – `POST /accounts`

**Request body** (per assignment):

```json
{
  "account_id": 123,
  "initial_balance": "100.23344"
}
```

- `account_id` – integer `BIGINT`, chosen by the caller.
- `initial_balance` – string representation of the starting balance.

**Example curl**:

```bash
curl -i -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 201, "initial_balance": "500.00"}'
```

**Successful response**:

- Status: `201 Created`
- Body: empty (per exercise spec)

**Error cases** (selected):

- `400 Bad Request` – invalid JSON, negative balance, non-positive ID.
- `409 Conflict` – account already exists.
- `500 Internal Server Error` – unexpected DB or service error.

### 5.2 Get account – `GET /accounts/{account_id}`

Returns the current balance for an account.

**Example curl**:

```bash
curl -i http://localhost:8080/accounts/201
```

**Successful response**:

```json
{
  "account_id": 201,
  "balance": "500.00000"
}
```

**Error cases**:

- `400 Bad Request` – invalid account ID format.
- `404 Not Found` – account does not exist.

### 5.3 Submit transfer – `POST /transactions`

**Request body** (per assignment):

```json
{
  "source_account_id": 123,
  "destination_account_id": 456,
  "amount": "100.12345"
}
```

**Example curl**:

```bash
curl -i -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "100.00"}'
```

**Successful response**:

- Status: `201 Created`
- Body (informational only):

```json
{
  "message": "Transfer completed successfully",
  "amount": "100.00",
  "from": "201",
  "to": "202"
}
```

**Error mapping** (via `TransferError` / `AccountError`):

- `400 Bad Request` – same-account transfer, non-positive amount, invalid IDs,
  missing fields, invalid JSON.
- `404 Not Found` – source or destination account does not exist.
- `422 Unprocessable Entity` – insufficient funds.
- `500 Internal Server Error` – unexpected DB/service failures.

---

## 6. Concurrency & Data Integrity

Concurrency is handled in the **service layer** using a combination of:

- **Serializable isolation** – each transfer runs inside a transaction created
  with `sql.LevelSerializable`, the strongest isolation level available.
- **Row-level locking** – balances are accessed via `SELECT ... FOR UPDATE`
  (`AccountRepository.GetByIDWithLock`), ensuring no two transfers modify the
  same account row concurrently without serialisation.
- **Non-negative balance invariant** – `UpdateBalanceTx` computes the new
  balance in memory and rejects any operation that would result in a negative
  balance.
- **Atomic updates** – debiting, crediting, and inserting into `transactions`
  are performed in the same DB transaction.

---


## 7. Assumptions & Notes

- Single currency across all accounts.
- No authentication or authorisation (as per the assignment).
- All monetary amounts are provided and returned as **strings** with up to five
  decimal places to match the exercise examples; the service uses
  `shopspring/decimal` internally and the database stores values in
  `DECIMAL(20,5)`.
- Account IDs are `BIGINT`, chosen by the caller; the system does not generate
  IDs.
- Error responses are simple plain-text messages with appropriate HTTP codes
  (sufficient for this exercise, but could be extended to structured JSON
  error objects).

---

## 8. Extensibility

The current design is intentionally modular so extensions are straightforward:

- New endpoints (e.g. transaction history) can be added in `internal/api`
  backed by service methods in `internal/service`.
- Additional invariants (e.g. transfer limits, daily caps) belong in the
  service layer.
- More sophisticated logging/metrics backends can be plugged in behind the
  `utils.Logger` abstraction.

This makes the implementation suitable not only for the current assignment
but also as a foundation for further interview discussion around design and
trade-offs.
