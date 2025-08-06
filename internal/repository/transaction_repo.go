package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/hidimpu/transfersystem/internal/model"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateTransaction creates a new transaction record with proper locking
func (r *TransactionRepository) CreateTransaction(ctx context.Context, txn *model.Transaction, tx *sql.Tx) error {
	query := `
        INSERT INTO transactions (source_account_id, destination_account_id, amount, created_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `
	err := tx.QueryRowContext(
		ctx,
		query,
		txn.SourceAccountID,
		txn.DestinationAccountID,
		txn.Amount,
		time.Now(),
	).Scan(&txn.ID)

	return err
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*model.Transaction, error) {
	var txn model.Transaction
	err := r.db.QueryRowContext(ctx, `
		SELECT id, source_account_id, destination_account_id, amount, created_at 
		FROM transactions WHERE id = $1`, id).Scan(
		&txn.ID, &txn.SourceAccountID, &txn.DestinationAccountID, &txn.Amount, &txn.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &txn, nil
}

// GetByAccountID retrieves all transactions for a specific account
func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID int64) ([]*model.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_account_id, destination_account_id, amount, created_at 
		FROM transactions 
		WHERE source_account_id = $1 OR destination_account_id = $1 
		ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		if err := rows.Scan(&txn.ID, &txn.SourceAccountID, &txn.DestinationAccountID, &txn.Amount, &txn.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &txn)
	}
	return transactions, nil
}

// GetAll retrieves all transactions (for admin purposes)
func (r *TransactionRepository) GetAll(ctx context.Context) ([]*model.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_account_id, destination_account_id, amount, created_at 
		FROM transactions 
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		if err := rows.Scan(&txn.ID, &txn.SourceAccountID, &txn.DestinationAccountID, &txn.Amount, &txn.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &txn)
	}
	return transactions, nil
}

// GetTransactionHistory retrieves transaction history with pagination
func (r *TransactionRepository) GetTransactionHistory(ctx context.Context, accountID int64, limit, offset int) ([]*model.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_account_id, destination_account_id, amount, created_at 
		FROM transactions 
		WHERE source_account_id = $1 OR destination_account_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`, accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		var txn model.Transaction
		if err := rows.Scan(&txn.ID, &txn.SourceAccountID, &txn.DestinationAccountID, &txn.Amount, &txn.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, &txn)
	}
	return transactions, nil
}
