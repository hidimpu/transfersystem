package model

import "errors"

var ErrInsufficientBalance = errors.New("insufficient balance")

type Account struct {
	AccountID      int     `json:"account_id"`
	InitialBalance float64 `json:"balance"`
}

type Transaction struct {
	SourceAccountID      int     `json:"source_account_id"`
	DestinationAccountID int     `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}
