package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"transfersystem/internal/model"

	"github.com/go-chi/chi/v5"
)

type DB interface {
	CreateAccount(*model.Account) error
	GetAccount(id int) (*model.Account, error)
	ProcessTransaction(*model.Transaction) error
}

type Handler struct {
	DB DB
}

func NewHandler(db DB) *Handler {
	return &Handler{DB: db}
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func HandleJSONDecode(err error, w http.ResponseWriter) bool {
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return true
	}
	return false
}

func (h *Handler) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "only POST allowed")
		return
	}

	var acc model.Account
	if HandleJSONDecode(json.NewDecoder(r.Body).Decode(&acc), w) {
		return
	}

	if acc.AccountID <= 0 || acc.InitialBalance < 0 {
		respondError(w, http.StatusBadRequest, "invalid account input")
		return
	}

	if err := h.DB.CreateAccount(&acc); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "account_id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "invalid account ID")
		return
	}

	acc, err := h.DB.GetAccount(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}

	json.NewEncoder(w).Encode(acc)
}

func (h *Handler) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	var tx model.Transaction
	if HandleJSONDecode(json.NewDecoder(r.Body).Decode(&tx), w) {
		return
	}

	if tx.SourceAccountID <= 0 || tx.DestinationAccountID <= 0 || tx.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "invalid transaction input")
		return
	}

	if tx.SourceAccountID == tx.DestinationAccountID {
		respondError(w, http.StatusUnprocessableEntity, "source and destination must differ")
		return
	}

	if err := h.DB.ProcessTransaction(&tx); err != nil {
		if errors.Is(err, model.ErrInsufficientBalance) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
