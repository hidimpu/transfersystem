#!/usr/bin/env bash
set -euo pipefail

# Simple end-to-end test script for the transfersystem API.
#
# It will:
#   1. Run Go unit tests
#   2. Start the API server
#   3. Exercise the HTTP endpoints with a variety of scenarios
#   4. Check HTTP status codes and basic balance correctness
#
# Requirements:
#   - go
#   - curl
#   - jq (for JSON parsing)
#   - A running PostgreSQL instance reachable via DB_URL
#
# Usage:
#   From the repository root:
#     chmod +x ./test_scenarios.sh
#     ./test_scenarios.sh

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

if ! command -v go >/dev/null 2>&1; then
  echo "[ERROR] go is required but not installed" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "[ERROR] curl is required but not installed" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "[ERROR] jq is required but not installed" >&2
  exit 1
fi

BASE_URL="${BASE_URL:-http://localhost:8080}"
DB_URL_ENV="${DB_URL:-postgres://postgres:postgres@localhost:5432/transfersystem?sslmode=disable}"
PORT_ENV="${PORT:-8080}"

log() {
  echo "[$(date +"%H:%M:%S")] $*"
}

fail() {
  echo "[FAIL] $*" >&2
  exit 1
}

assert_http_code() {
  local name="$1"; shift
  local expected="$1"; shift
  local cmd=("$@")

  local tmp_body
  tmp_body="$(mktemp)"

  # shellcheck disable=SC2068
  local status
  status=$(${cmd[@]} -s -o "$tmp_body" -w '%{http_code}')

  if [[ "$status" != "$expected" ]]; then
    echo "[FAIL] $name: expected HTTP $expected, got $status" >&2
    echo "Response body:" >&2
    cat "$tmp_body" >&2
    rm -f "$tmp_body"
    exit 1
  fi

  echo "[OK] $name (HTTP $status)"
  rm -f "$tmp_body"
}

# Allow minor formatting differences between e.g. "1000.00" and "1000"
normalize_amount() {
  local amt="$1"
  # Trim surrounding quotes or spaces
  amt="${amt%\"}"
  amt="${amt#\"}"
  amt="${amt## }"
  amt="${amt%% }"
  echo "$amt"
}

assert_balance() {
  local account_id="$1"
  local expected_raw="$2"  # e.g. 1000.00

  local resp
  resp="$(curl -s "${BASE_URL}/accounts/${account_id}")" || fail "Failed to GET account ${account_id}"

  local actual
  actual="$(echo "$resp" | jq -r '.balance')"
  actual="$(normalize_amount "$actual")"

  # Accept both exact match and trimmed ".00" variant
  local expected
  expected="$(normalize_amount "$expected_raw")"
  local alt="${expected}"
  if [[ "$expected" == *".00" ]]; then
    alt="${expected%.*}"
  fi

  if [[ "$actual" == "$expected" || "$actual" == "$alt" ]]; then
    echo "[OK] Balance for account ${account_id} is ${actual} (expected ~ ${expected})"
  else
    echo "[FAIL] Balance for account ${account_id}: expected ${expected} (or ${alt}), got ${actual}" >&2
    echo "Full response: $resp" >&2
    exit 1
  fi
}

start_server() {
  log "Starting API server on port ${PORT_ENV}..."
  (
    export DB_URL="$DB_URL_ENV"
    export PORT="$PORT_ENV"
    cd "$REPO_ROOT/transfersystem" 2>/dev/null || cd "$REPO_ROOT"
    go run ./cmd/main.go
  ) &
  SERVER_PID=$!
  log "Server PID: $SERVER_PID"
  # Give the server a moment to start
  sleep 3
}

stop_server() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    log "Stopping API server (PID $SERVER_PID)..."
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}

trap stop_server EXIT

log "Running Go unit tests..."
cd "$REPO_ROOT/transfersystem" 2>/dev/null || true

go test ./... >/dev/null
log "Unit tests passed."

start_server

# Generate unique account IDs for this run to avoid clashes
RUN_ID="$(date +%s)"
ACC_A=$((RUN_ID + 1))
ACC_B=$((RUN_ID + 2))
ACC_SRC_CONC=$((RUN_ID + 3))
ACC_DST_CONC=$((RUN_ID + 4))

log "Using account IDs: A=${ACC_A}, B=${ACC_B}, SRC_CONC=${ACC_SRC_CONC}, DST_CONC=${ACC_DST_CONC}"

log "=== Scenario 1: Create accounts ==="
assert_http_code "Create account A" 201 curl -X POST "${BASE_URL}/accounts" \
  -H 'Content-Type: application/json' \
  -d "{\"account_id\": ${ACC_A}, \"initial_balance\": \"1000.00\"}"

assert_http_code "Create account B" 201 curl -X POST "${BASE_URL}/accounts" \
  -H 'Content-Type: application/json' \
  -d "{\"account_id\": ${ACC_B}, \"initial_balance\": \"500.00\"}"

# Duplicate account creation should fail with conflict
assert_http_code "Duplicate create account A" 409 curl -X POST "${BASE_URL}/accounts" \
  -H 'Content-Type: application/json' \
  -d "{\"account_id\": ${ACC_A}, \"initial_balance\": \"1000.00\"}"

log "=== Scenario 2: Get account balances ==="
assert_http_code "Get account A" 200 curl "${BASE_URL}/accounts/${ACC_A}"
assert_http_code "Get account B" 200 curl "${BASE_URL}/accounts/${ACC_B}"

assert_balance "$ACC_A" "1000.00"
assert_balance "$ACC_B" "500.00"

log "=== Scenario 3: Valid transfer A -> B (100.00) ==="
assert_http_code "Transfer 100 from A to B" 201 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}, \"amount\": \"100.00\"}"

assert_balance "$ACC_A" "900.00"
assert_balance "$ACC_B" "600.00"

log "=== Scenario 4: Insufficient funds ==="
assert_http_code "Transfer too much from A to B" 422 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}, \"amount\": \"10000.00\"}"

log "=== Scenario 5: Same-account transfer ==="
assert_http_code "Transfer A -> A" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_A}, \"amount\": \"10.00\"}"

log "=== Scenario 6: Negative and zero amounts ==="
assert_http_code "Negative amount" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}, \"amount\": \"-5.00\"}"

assert_http_code "Zero amount" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}, \"amount\": \"0.00\"}"

log "=== Scenario 7: Non-existent accounts ==="
NON_EXISTENT_SRC=$((RUN_ID + 100000))
NON_EXISTENT_DST=$((RUN_ID + 200000))

assert_http_code "Non-existent source" 404 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${NON_EXISTENT_SRC}, \"destination_account_id\": ${ACC_B}, \"amount\": \"10.00\"}"

assert_http_code "Non-existent destination" 404 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${NON_EXISTENT_DST}, \"amount\": \"10.00\"}"

log "=== Scenario 8: Invalid payloads ==="
# Missing amount
assert_http_code "Missing amount field" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}}"

# Invalid amount format
assert_http_code "Invalid amount format" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"destination_account_id\": ${ACC_B}, \"amount\": \"abc\"}"

# Missing source_account_id
assert_http_code "Missing source_account_id" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"destination_account_id\": ${ACC_B}, \"amount\": \"10.00\"}"

# Missing destination_account_id
assert_http_code "Missing destination_account_id" 400 curl -X POST "${BASE_URL}/transactions" \
  -H 'Content-Type: application/json' \
  -d "{\"source_account_id\": ${ACC_A}, \"amount\": \"10.00\"}"

log "=== Scenario 9: Concurrency test (multiple parallel transfers) ==="

# Create two fresh accounts specifically for the concurrency test
assert_http_code "Create SRC_CONC" 201 curl -X POST "${BASE_URL}/accounts" \
  -H 'Content-Type: application/json' \
  -d "{\"account_id\": ${ACC_SRC_CONC}, \"initial_balance\": \"100.00\"}"

assert_http_code "Create DST_CONC" 201 curl -X POST "${BASE_URL}/accounts" \
  -H 'Content-Type: application/json' \
  -d "{\"account_id\": ${ACC_DST_CONC}, \"initial_balance\": \"0.00\"}"

# Fire 10 parallel transfers of 1.00 from SRC_CONC to DST_CONC
TRANSFERS=10
for i in $(seq 1 "$TRANSFERS"); do
  curl -s -o /dev/null -w '' -X POST "${BASE_URL}/transactions" \
    -H 'Content-Type: application/json' \
    -d "{\"source_account_id\": ${ACC_SRC_CONC}, \"destination_account_id\": ${ACC_DST_CONC}, \"amount\": \"1.00\"}" &
done
wait

# Expect SRC_CONC: 100 - 10 = 90, DST_CONC: 0 + 10 = 10
assert_balance "$ACC_SRC_CONC" "90.00"
assert_balance "$ACC_DST_CONC" "10.00"

log "All scenarios completed successfully."
