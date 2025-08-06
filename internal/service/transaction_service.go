package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hidimpu/transfersystem/internal/repository"

	"github.com/shopspring/decimal"
)

type TransactionService struct {
	db      *sql.DB
	accRepo *repository.AccountRepository
}

func NewTransactionService(db *sql.DB, accRepo *repository.AccountRepository) *TransactionService {
	return &TransactionService{db: db, accRepo: accRepo}
}

// Transfer handles concurrency and atomicity via DB transactions and row locks.
func (s *TransactionService) Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
	if srcID == dstID {
		return fmt.Errorf("cannot transfer to same account")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("transfer amount must be positive")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Debit source, atomically, with row locked.
	err = s.accRepo.UpdateBalanceTx(ctx, srcID, amount.Neg(), tx)
	if err != nil {
		return err
	}
	// Credit destination, with row locked.
	err = s.accRepo.UpdateBalanceTx(ctx, dstID, amount, tx)
	if err != nil {
		return err
	}
	// Record transaction.
	_, err = tx.ExecContext(ctx, `INSERT INTO transactions (source_account_id, destination_account_id, amount) VALUES ($1, $2, $3)`, srcID, dstID, amount)
	return err
}
