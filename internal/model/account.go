package model

import "github.com/shopspring/decimal"

type Account struct {
	ID      int64           `json:"account_id"`
	Balance decimal.Decimal `json:"balance"`
}
