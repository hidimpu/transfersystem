package model

import "github.com/shopspring/decimal"

type Transaction struct {
    SourceAccountID      int64           `json:"source_account_id"`
    DestinationAccountID int64           `json:"destination_account_id"`
    Amount               decimal.Decimal `json:"amount"`
}
