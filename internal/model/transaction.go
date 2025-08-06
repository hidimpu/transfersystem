package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID                   int64           `json:"id"`
	SourceAccountID      int64           `json:"source_account_id"`
	DestinationAccountID int64           `json:"destination_account_id"`
	Amount               decimal.Decimal `json:"amount"`
	CreatedAt            time.Time       `json:"created_at"`
}
