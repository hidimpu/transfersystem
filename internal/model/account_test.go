package model

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAccount_Validation(t *testing.T) {
	tests := []struct {
		name    string
		account Account
		isValid bool
	}{
		{
			name: "valid account",
			account: Account{
				ID:      1,
				Balance: decimal.NewFromFloat(1000.00),
			},
			isValid: true,
		},
		{
			name: "account with zero ID",
			account: Account{
				ID:      0,
				Balance: decimal.NewFromFloat(100.00),
			},
			isValid: false,
		},
		{
			name: "account with negative ID",
			account: Account{
				ID:      -1,
				Balance: decimal.NewFromFloat(100.00),
			},
			isValid: false,
		},
		{
			name: "account with negative balance",
			account: Account{
				ID:      1,
				Balance: decimal.NewFromFloat(-100.00),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.account.ID > 0 && !tt.account.Balance.IsNegative()
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestAccount_BalanceOperations(t *testing.T) {
	account := Account{
		ID:      1,
		Balance: decimal.NewFromFloat(1000.00),
	}

	t.Run("add to balance", func(t *testing.T) {
		amount := decimal.NewFromFloat(100.00)
		newBalance := account.Balance.Add(amount)
		expected := decimal.NewFromFloat(1100.00)
		assert.True(t, newBalance.Equal(expected))
	})

	t.Run("subtract from balance", func(t *testing.T) {
		amount := decimal.NewFromFloat(200.00)
		newBalance := account.Balance.Sub(amount)
		expected := decimal.NewFromFloat(800.00)
		assert.True(t, newBalance.Equal(expected))
	})

	t.Run("insufficient funds check", func(t *testing.T) {
		amount := decimal.NewFromFloat(1500.00)
		newBalance := account.Balance.Sub(amount)
		assert.True(t, newBalance.IsNegative())
	})
}
