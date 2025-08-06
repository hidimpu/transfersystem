package service

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestValidateTransferRequest(t *testing.T) {
	tests := []struct {
		name        string
		srcID       int64
		dstID       int64
		amount      decimal.Decimal
		expectedErr string
	}{
		{
			name:        "valid transfer",
			srcID:       1,
			dstID:       2,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "",
		},
		{
			name:        "same account transfer",
			srcID:       1,
			dstID:       1,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "cannot transfer to same account",
		},
		{
			name:        "negative amount",
			srcID:       1,
			dstID:       2,
			amount:      decimal.NewFromFloat(-100.00),
			expectedErr: "transfer amount must be positive",
		},
		{
			name:        "zero amount",
			srcID:       1,
			dstID:       2,
			amount:      decimal.Zero,
			expectedErr: "transfer amount must be positive",
		},
		{
			name:        "invalid source account ID",
			srcID:       0,
			dstID:       2,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "invalid account IDs",
		},
		{
			name:        "invalid destination account ID",
			srcID:       1,
			dstID:       0,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "invalid account IDs",
		},
		{
			name:        "negative source account ID",
			srcID:       -1,
			dstID:       2,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "invalid account IDs",
		},
		{
			name:        "negative destination account ID",
			srcID:       1,
			dstID:       -2,
			amount:      decimal.NewFromFloat(100.00),
			expectedErr: "invalid account IDs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransferRequest(tt.srcID, tt.dstID, tt.amount)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAccount(t *testing.T) {
	tests := []struct {
		name        string
		accountID   int64
		balance     decimal.Decimal
		expectedErr string
	}{
		{
			name:        "valid account",
			accountID:   1,
			balance:     decimal.NewFromFloat(1000.00),
			expectedErr: "",
		},
		{
			name:        "zero account ID",
			accountID:   0,
			balance:     decimal.NewFromFloat(1000.00),
			expectedErr: "account ID must be positive",
		},
		{
			name:        "negative account ID",
			accountID:   -1,
			balance:     decimal.NewFromFloat(1000.00),
			expectedErr: "account ID must be positive",
		},
		{
			name:        "negative balance",
			accountID:   1,
			balance:     decimal.NewFromFloat(-100.00),
			expectedErr: "account balance cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAccount(tt.accountID, tt.balance)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper functions for testing
func validateTransferRequest(srcID, dstID int64, amount decimal.Decimal) error {
	if srcID == dstID {
		return fmt.Errorf("cannot transfer to same account")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be positive")
	}
	if srcID <= 0 || dstID <= 0 {
		return fmt.Errorf("invalid account IDs")
	}
	return nil
}

func validateAccount(accountID int64, balance decimal.Decimal) error {
	if accountID <= 0 {
		return fmt.Errorf("account ID must be positive")
	}
	if balance.IsNegative() {
		return fmt.Errorf("account balance cannot be negative")
	}
	return nil
}
