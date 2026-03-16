package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hidimpu/transfersystem/internal/model"
	"github.com/hidimpu/transfersystem/internal/service"
	"github.com/hidimpu/transfersystem/internal/utils"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	service *service.TransactionService
	logger  *utils.Logger
}

func NewTransactionHandler(s *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service: s,
		logger:  utils.GlobalLogger,
	}
}

func (h *TransactionHandler) TransferFunds(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceAccountID      int64  `json:"source_account_id"`
		DestinationAccountID int64  `json:"destination_account_id"`
		Amount               string `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogError("API_TRANSFER", "JSON_DECODE_ERROR", "Invalid JSON payload", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SourceAccountID == 0 {
		h.logger.LogWarning("API_TRANSFER", "Missing source_account_id")
		http.Error(w, "source_account_id is required", http.StatusBadRequest)
		return
	}
	if req.DestinationAccountID == 0 {
		h.logger.LogWarning("API_TRANSFER", "Missing destination_account_id")
		http.Error(w, "destination_account_id is required", http.StatusBadRequest)
		return
	}
	if req.Amount == "" {
		h.logger.LogWarning("API_TRANSFER", "Missing amount")
		http.Error(w, "amount is required", http.StatusBadRequest)
		return
	}

	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		h.logger.LogError("API_TRANSFER", "AMOUNT_PARSE_ERROR", "Invalid amount format", err)
		http.Error(w, "Invalid amount format", http.StatusBadRequest)
		return
	}

	if err := h.service.Transfer(r.Context(), req.SourceAccountID, req.DestinationAccountID, amt); err != nil {
		// Map service errors to appropriate HTTP status codes using error enums
		var statusCode int
		var errorMessage string

		switch err := err.(type) {
		case model.TransferError:
			statusCode = err.HTTPStatus()
			errorMessage = err.Error()
		case model.AccountError:
			statusCode = err.HTTPStatus()
			errorMessage = err.Error()
		default:
			statusCode = http.StatusInternalServerError
			errorMessage = "Internal server error"
		}

		h.logger.LogError("API_TRANSFER", "TRANSFER_ERROR", errorMessage, err)
		http.Error(w, errorMessage, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transfer completed successfully",
		"amount":  amt.String(),
		"from":    fmt.Sprintf("%d", req.SourceAccountID),
		"to":      fmt.Sprintf("%d", req.DestinationAccountID),
	})
}
