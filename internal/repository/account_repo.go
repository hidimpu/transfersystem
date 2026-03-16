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

// Create creates a new account with proper validation
func (r *AccountRepository) Create(ctx context.Context, acc *model.Account) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO accounts(account_id, balance) VALUES($1, $2)`, acc.ID, acc.Balance)
	return err
}

// GetByID retrieves an account by ID with proper error handling
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*model.Account, error) {
	var acc model.Account
	var balanceStr string
	err := r.db.QueryRowContext(ctx, `SELECT account_id, balance FROM accounts WHERE account_id=$1`, id).Scan(&acc.ID, &balanceStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	acc.Balance, _ = decimal.NewFromString(balanceStr)
	return &acc, nil
}

// GetByIDWithLock retrieves an account with FOR UPDATE lock for transactions
func (r *AccountRepository) GetByIDWithLock(ctx context.Context, id int64, tx *sql.Tx) (*model.Account, error) {
	var acc model.Account
	var balanceStr string
	err := tx.QueryRowContext(ctx, `SELECT account_id, balance FROM accounts WHERE account_id=$1 FOR UPDATE`, id).Scan(&acc.ID, &balanceStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	acc.Balance, _ = decimal.NewFromString(balanceStr)
	return &acc, nil
}

// UpdateBalanceTx updates account balance with row-level locking and validation
func (r *AccountRepository) UpdateBalanceTx(ctx context.Context, id int64, diff decimal.Decimal, tx *sql.Tx) error {
	// First, get the account with lock
	account, err := r.GetByIDWithLock(ctx, id, tx)
	if err != nil {
		return err
	}

	// Calculate new balance
	newBalance := account.Balance.Add(diff)
	if newBalance.IsNegative() {
		return fmt.Errorf("insufficient funds")
	}

	// Update the balance
	_, err = tx.ExecContext(ctx, `UPDATE accounts SET balance=$1 WHERE account_id=$2`, newBalance, id)
	return err
}

// GetAll retrieves all accounts (for admin purposes)
func (r *AccountRepository) GetAll(ctx context.Context) ([]*model.Account, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT account_id, balance FROM accounts ORDER BY account_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*model.Account
	for rows.Next() {
		var acc model.Account
		var balanceStr string
		if err := rows.Scan(&acc.ID, &balanceStr); err != nil {
			return nil, err
		}
		acc.Balance, _ = decimal.NewFromString(balanceStr)
		accounts = append(accounts, &acc)
	}
	return accounts, nil
}

// Exists checks if an account exists
func (r *AccountRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_id=$1)`, id).Scan(&exists)
	return exists, err
}

// GetBalance retrieves only the balance of an account (optimized for frequent checks)
func (r *AccountRepository) GetBalance(ctx context.Context, id int64) (decimal.Decimal, error) {
	var balanceStr string
	err := r.db.QueryRowContext(ctx, `SELECT balance FROM accounts WHERE account_id=$1`, id).Scan(&balanceStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return decimal.Zero, fmt.Errorf("account not found")
		}
		return decimal.Zero, err
	}
	return decimal.NewFromString(balanceStr)
}
