#!/usr/bin/env bash

# Simple script to hit all test endpoints and show responses.
# Usage:
#   chmod +x test_all.sh
#   ./test_all.sh
#
# Optionally:
#   BASE_URL="http://localhost:8081" ./test_all.sh

set -u

BASE_URL="${BASE_URL:-http://localhost:8080}"

run() {
  local name="$1"; shift
  echo
  echo "==================================================================="
  echo ">>> $name"
  echo "-------------------------------------------------------------------"
  echo "+ $*"
  echo "-------------------------------------------------------------------"
  # -i = include HTTP headers, -s = silent progress bar, -S = show errors
  curl -i -sS "$@"
  echo
  echo "==================================================================="
  echo
}

# 1) Create account A (201, 1000.00)
run "Create account A (201, 1000.00)" \
  -X POST "${BASE_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{"account_id": 201, "initial_balance": "1000.00"}'

# 2) Create account B (202, 500.00)
run "Create account B (202, 500.00)" \
  -X POST "${BASE_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{"account_id": 202, "initial_balance": "500.00"}'

# 3) Duplicate account A (should be 409 Conflict)
run "Duplicate account A (should be 409)" \
  -X POST "${BASE_URL}/accounts" \
  -H "Content-Type: application/json" \
  -d '{"account_id": 201, "initial_balance": "1000.00"}'

# 4) Get account A
run "Get account A (201)" \
  "${BASE_URL}/accounts/201"

# 5) Get account B
run "Get account B (202)" \
  "${BASE_URL}/accounts/202"

# 6) Valid transfer: 100.00 from A (201) to B (202)
run "Valid transfer 100.00 from 201 -> 202" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "100.00"}'

# 7) Check balances after transfer
run "Get account A (201) after transfer" \
  "${BASE_URL}/accounts/201"

run "Get account B (202) after transfer" \
  "${BASE_URL}/accounts/202"

# 8) Insufficient funds: huge transfer from A to B (should be 422)
run "Insufficient funds transfer 1000000.00 from 201 -> 202 (should be 422)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "1000000.00"}'

# 9) Same-account transfer (should be 400)
run "Same-account transfer 201 -> 201 (should be 400)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 201, "amount": "10.00"}'

# 10) Negative amount (should be 400)
run "Negative amount -5.00 (should be 400)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "-5.00"}'

# 11) Zero amount (should be 400)
run "Zero amount 0.00 (should be 400)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "0.00"}'

# 12) Non-existent source account (should be 404)
run "Non-existent source account 999999 -> 202 (should be 404)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 999999, "destination_account_id": 202, "amount": "10.00"}'

# 13) Non-existent destination account (should be 404)
run "Non-existent destination account 201 -> 999999 (should be 404)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 999999, "amount": "10.00"}'

# 14) Invalid payload: missing amount (should be 400)
run "Invalid payload: missing amount (should be 400)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202}'

# 15) Invalid payload: non-numeric amount (should be 400)
run "Invalid payload: non-numeric amount 'abc' (should be 400)" \
  -X POST "${BASE_URL}/transactions" \
  -H "Content-Type: application/json" \
  -d '{"source_account_id": 201, "destination_account_id": 202, "amount": "abc"}'

echo
echo "All requests executed."