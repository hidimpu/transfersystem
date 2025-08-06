package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hidimpu/transfersystem/internal/model"

	"github.com/shopspring/decimal"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *model.Account) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO accounts(account_id, balance) VALUES($1, $2)`, acc.AccountID, acc.Balance)
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	var acc model.Account
	var balanceStr string
	err := r.db.QueryRowContext(ctx, `SELECT account_id, balance FROM accounts WHERE account_id=$1`, id).Scan(&acc.AccountID, &balanceStr)
	if err != nil {
		return nil, err
	}
	acc.Balance, _ = decimal.NewFromString(balanceStr)
	return &acc, nil
}

// Now takes an explicit tx for locking and atomic updates
func (r *AccountRepository) UpdateBalanceTx(ctx context.Context, id int64, diff decimal.Decimal, tx *sql.Tx) error {
	row := tx.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE account_id=$1 FOR UPDATE`, id)
	var balanceStr string
	if err := row.Scan(&balanceStr); err != nil {
		return err
	}
	balance, _ := decimal.NewFromString(balanceStr)
	newBalance := balance.Add(diff)
	if newBalance.IsNegative() {
		return fmt.Errorf("insufficient funds")
	}
	_, err := tx.ExecContext(ctx, `UPDATE accounts SET balance=$1 WHERE account_id=$2`, newBalance, id)
	return err
}
