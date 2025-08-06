package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidimpu/transfersystem/internal/model"
	"github.com/shopspring/decimal"
)

func CreateAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var acc model.Account
		if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if acc.ID == 0 || acc.Balance.IsNegative() {
			http.Error(w, "Invalid account details", http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO accounts (id, balance) VALUES ($1, $2)", acc.ID, acc.Balance)
		if err != nil {
			http.Error(w, "Failed to create account", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func GetAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountIDStr := chi.URLParam(r, "account_id")
		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid account ID", http.StatusBadRequest)
			return
		}

		row := db.QueryRow("SELECT id, balance FROM accounts WHERE id = $1", accountID)
		var acc model.Account
		if err := row.Scan(&acc.ID, &acc.Balance); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Account not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve account", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(acc)
	}
}

func CreateTransactionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var txReq model.Transaction
		if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if txReq.SourceAccountID == 0 || txReq.DestinationAccountID == 0 || txReq.Amount.LessThanOrEqual(decimal.Zero) {
			http.Error(w, "Invalid transaction data", http.StatusBadRequest)
			return
		}

		txReq.Amount = txReq.Amount.Round(2)

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Deduct from source
		_, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, txReq.Amount, txReq.SourceAccountID)
		if err != nil {
			http.Error(w, "Failed to debit source account", http.StatusInternalServerError)
			return
		}

		// Add to destination
		_, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, txReq.Amount, txReq.DestinationAccountID)
		if err != nil {
			http.Error(w, "Failed to credit destination account", http.StatusInternalServerError)
			return
		}

		// Record transaction
		_, err = tx.Exec(`INSERT INTO transactions (source_account_id, destination_account_id, amount) VALUES ($1, $2, $3)`,
			txReq.SourceAccountID, txReq.DestinationAccountID, txReq.Amount)
		if err != nil {
			http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
			return
		}

		if err = tx.Commit(); err != nil {
			http.Error(w, "Transaction commit failed", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
