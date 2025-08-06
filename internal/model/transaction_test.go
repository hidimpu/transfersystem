package model

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Validation(t *testing.T) {
	tests := []struct {
		name        string
		transaction Transaction
		isValid     bool
	}{
		{
			name: "valid transaction",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               decimal.NewFromFloat(100.00),
				CreatedAt:            time.Now(),
			},
			isValid: true,
		},
		{
			name: "transaction with zero source account",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      0,
				DestinationAccountID: 2,
				Amount:               decimal.NewFromFloat(100.00),
				CreatedAt:            time.Now(),
			},
			isValid: false,
		},
		{
			name: "transaction with zero destination account",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      1,
				DestinationAccountID: 0,
				Amount:               decimal.NewFromFloat(100.00),
				CreatedAt:            time.Now(),
			},
			isValid: false,
		},
		{
			name: "transaction with negative amount",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               decimal.NewFromFloat(-100.00),
				CreatedAt:            time.Now(),
			},
			isValid: false,
		},
		{
			name: "transaction with zero amount",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               decimal.Zero,
				CreatedAt:            time.Now(),
			},
			isValid: false,
		},
		{
			name: "transaction to same account",
			transaction: Transaction{
				ID:                   1,
				SourceAccountID:      1,
				DestinationAccountID: 1,
				Amount:               decimal.NewFromFloat(100.00),
				CreatedAt:            time.Now(),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.transaction.SourceAccountID > 0 &&
				tt.transaction.DestinationAccountID > 0 &&
				tt.transaction.SourceAccountID != tt.transaction.DestinationAccountID &&
				tt.transaction.Amount.GreaterThan(decimal.Zero)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestTransaction_AmountOperations(t *testing.T) {
	transaction := Transaction{
		ID:                   1,
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromFloat(100.00),
		CreatedAt:            time.Now(),
	}

	t.Run("amount is positive", func(t *testing.T) {
		assert.True(t, transaction.Amount.GreaterThan(decimal.Zero))
	})

	t.Run("amount is not negative", func(t *testing.T) {
		assert.False(t, transaction.Amount.LessThan(decimal.Zero))
	})

	t.Run("amount is not zero", func(t *testing.T) {
		assert.False(t, transaction.Amount.Equal(decimal.Zero))
	})
}

func TestTransaction_DifferentAccounts(t *testing.T) {
	t.Run("different source and destination accounts", func(t *testing.T) {
		transaction := Transaction{
			ID:                   1,
			SourceAccountID:      1,
			DestinationAccountID: 2,
			Amount:               decimal.NewFromFloat(100.00),
			CreatedAt:            time.Now(),
		}
		assert.NotEqual(t, transaction.SourceAccountID, transaction.DestinationAccountID)
	})

	t.Run("same source and destination accounts", func(t *testing.T) {
		transaction := Transaction{
			ID:                   1,
			SourceAccountID:      1,
			DestinationAccountID: 1,
			Amount:               decimal.NewFromFloat(100.00),
			CreatedAt:            time.Now(),
		}
		assert.Equal(t, transaction.SourceAccountID, transaction.DestinationAccountID)
	})
}
