package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidimpu/transfersystem/internal/model"
	"github.com/hidimpu/transfersystem/internal/service"
	"github.com/hidimpu/transfersystem/internal/utils"
	"github.com/shopspring/decimal"
)

// Service-based handlers (NEW)
//
// Specification alignment:
//   - Request body: {"account_id": 123, "initial_balance": "100.23344"}
//   - Response: on success, an empty body with appropriate status code.
func CreateAccountServiceHandler(accountService service.AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := utils.GlobalLogger

		// Define a lightweight DTO that matches the exercise specification.
		var req struct {
			AccountID      int64  `json:"account_id"`
			InitialBalance string `json:"initial_balance"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.LogError("API_ACCOUNT_CREATE", "JSON_DECODE_ERROR", "Invalid JSON payload", err)
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.InitialBalance == "" {
			logger.LogWarning("API_ACCOUNT_CREATE", "Missing initial_balance")
			http.Error(w, "initial_balance is required", http.StatusBadRequest)
			return
		}

		balance, err := decimal.NewFromString(req.InitialBalance)
		if err != nil {
			logger.LogError("API_ACCOUNT_CREATE", "BALANCE_PARSE_ERROR", "Invalid initial_balance format", err)
			http.Error(w, "Invalid initial_balance format", http.StatusBadRequest)
			return
		}

		acc := model.Account{
			ID:      req.AccountID,
			Balance: balance,
		}

		if err := accountService.CreateAccount(r.Context(), &acc); err != nil {
			// Map service errors to appropriate HTTP status codes using error enums
			var statusCode int
			var errorMessage string

			switch err := err.(type) {
			case model.AccountError:
				statusCode = err.HTTPStatus()
				errorMessage = err.Error()
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to create account"
			}

			logger.LogError("API_ACCOUNT_CREATE", "CREATE_ERROR", errorMessage, err)
			http.Error(w, errorMessage, statusCode)
			return
		}

		// Per exercise: return an empty successful response body.
		w.WriteHeader(http.StatusCreated)
	}
}

func GetAccountServiceHandler(accountService service.AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := utils.GlobalLogger

		accountIDStr := chi.URLParam(r, "account_id")
		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			logger.LogError("API_ACCOUNT_GET", "PARSE_ERROR", "Invalid account ID format", err)
			http.Error(w, "Invalid account ID format", http.StatusBadRequest)
			return
		}

		acc, err := accountService.GetAccountByID(r.Context(), accountID)
		if err != nil {
			// Map service errors to appropriate HTTP status codes using error enums
			var statusCode int
			var errorMessage string

			switch err := err.(type) {
			case model.AccountError:
				statusCode = err.HTTPStatus()
				errorMessage = err.Error()
			default:
				statusCode = http.StatusInternalServerError
				errorMessage = "Failed to retrieve account"
			}

			logger.LogError("API_ACCOUNT_GET", "GET_ERROR", errorMessage, err)
			http.Error(w, errorMessage, statusCode)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(acc)
	}
}
