package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hidimpu/transfersystem/internal/model"
	"github.com/hidimpu/transfersystem/internal/repository"
	"github.com/hidimpu/transfersystem/internal/utils"

	"github.com/shopspring/decimal"
)

type TransactionService struct {
	db          *sql.DB
	accountRepo *repository.AccountRepository
	txnRepo     *repository.TransactionRepository
	logger      *utils.Logger
}

func NewTransactionService(db *sql.DB, accRepo *repository.AccountRepository, txnRepo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{
		db:          db,
		accountRepo: accRepo,
		txnRepo:     txnRepo,
		logger:      utils.GlobalLogger,
	}
}

// Transfer handles concurrency and atomicity via DB transactions and row locks.
func (s *TransactionService) Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
	// Log transfer attempt
	s.logger.LogTransfer("ATTEMPT", srcID, dstID, amount.String(), false)

	// Business logic validation
	if err := s.validateTransferRequest(srcID, dstID, amount); err != nil {
		s.logger.LogError("TRANSFER_VALIDATION", "VALIDATION_ERROR", err.Error(), err)
		return err
	}

	// Start database transaction with proper isolation
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Highest isolation level for financial transactions
		ReadOnly:  false,
	})
	if err != nil {
		s.logger.LogError("TRANSFER_DB", "DB_CONNECTION_ERROR", "Failed to begin transaction", err)
		return model.ErrServiceUnavailable
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			s.logger.LogError("TRANSFER_DB", "TRANSACTION_ROLLBACK", "Transaction rolled back", err)
		} else {
			tx.Commit()
			s.logger.LogInfo("TRANSFER_DB", "Transaction committed successfully")
		}
	}()

	// Debit source account with row-level locking
	if err = s.accountRepo.UpdateBalanceTx(ctx, srcID, amount.Neg(), tx); err != nil {
		if strings.Contains(err.Error(), "insufficient funds") {
			s.logger.LogError("TRANSFER_DEBIT", "INSUFFICIENT_FUNDS", fmt.Sprintf("Account %d has insufficient funds", srcID), err)
			return model.ErrInsufficientFunds
		}
		s.logger.LogError("TRANSFER_DEBIT", "DEBIT_ERROR", fmt.Sprintf("Failed to debit account %d", srcID), err)
		return model.ErrFailedDebit
	}

	// Credit destination account with row-level locking
	if err = s.accountRepo.UpdateBalanceTx(ctx, dstID, amount, tx); err != nil {
		s.logger.LogError("TRANSFER_CREDIT", "CREDIT_ERROR", fmt.Sprintf("Failed to credit account %d", dstID), err)
		return model.ErrFailedCredit
	}

	// Record the transaction
	txn := &model.Transaction{
		SourceAccountID:      srcID,
		DestinationAccountID: dstID,
		Amount:               amount,
	}
	if err = s.txnRepo.CreateTransaction(ctx, txn, tx); err != nil {
		s.logger.LogError("TRANSFER_RECORD", "RECORD_ERROR", "Failed to record transaction", err)
		return model.ErrFailedRecordTxn
	}

	// Log successful transfer
	s.logger.LogTransfer("SUCCESS", srcID, dstID, amount.String(), true)
	return nil
}

// validateTransferRequest validates the transfer request before processing
func (s *TransactionService) validateTransferRequest(srcID, dstID int64, amount decimal.Decimal) error {
	// Check same account transfer FIRST (no database dependency)
	if srcID == dstID {
		s.logger.LogWarning("TRANSFER_VALIDATION", fmt.Sprintf("Same account transfer attempted: %d -> %d", srcID, dstID))
		return model.ErrSameAccountTransfer
	}

	// Check amount validation (no database dependency)
	if amount.LessThanOrEqual(decimal.Zero) {
		s.logger.LogWarning("TRANSFER_VALIDATION", fmt.Sprintf("Invalid amount: %s", amount.String()))
		return model.ErrNegativeAmount
	}

	// Check account ID validation (no database dependency)
	if srcID <= 0 || dstID <= 0 {
		s.logger.LogWarning("TRANSFER_VALIDATION", fmt.Sprintf("Invalid account IDs: src=%d, dst=%d", srcID, dstID))
		return model.ErrInvalidAccountIDs
	}

	// Check if both accounts exist - handle database errors gracefully
	srcExists, err := s.accountRepo.Exists(context.Background(), srcID)
	if err != nil {
		s.logger.LogError("TRANSFER_VALIDATION", "DB_ERROR", fmt.Sprintf("Failed to validate source account %d", srcID), err)
		return model.ErrSourceAccountNotFound
	}
	if !srcExists {
		s.logger.LogWarning("TRANSFER_VALIDATION", fmt.Sprintf("Source account not found: %d", srcID))
		return model.ErrSourceAccountNotFound
	}

	dstExists, err := s.accountRepo.Exists(context.Background(), dstID)
	if err != nil {
		s.logger.LogError("TRANSFER_VALIDATION", "DB_ERROR", fmt.Sprintf("Failed to validate destination account %d", dstID), err)
		return model.ErrDestAccountNotFound
	}
	if !dstExists {
		s.logger.LogWarning("TRANSFER_VALIDATION", fmt.Sprintf("Destination account not found: %d", dstID))
		return model.ErrDestAccountNotFound
	}

	return nil
}

// GetTransactionHistory retrieves transaction history for an account
func (s *TransactionService) GetTransactionHistory(ctx context.Context, accountID int64, limit, offset int) ([]*model.Transaction, error) {
	if accountID <= 0 {
		return nil, fmt.Errorf("invalid account ID")
	}

	// Check if account exists
	exists, err := s.accountRepo.Exists(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate account: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	return s.txnRepo.GetTransactionHistory(ctx, accountID, limit, offset)
}

// GetTransactionByID retrieves a specific transaction
func (s *TransactionService) GetTransactionByID(ctx context.Context, transactionID int64) (*model.Transaction, error) {
	if transactionID <= 0 {
		return nil, fmt.Errorf("invalid transaction ID")
	}

	return s.txnRepo.GetByID(ctx, transactionID)
}

// GetAllTransactions retrieves all transactions (admin function)
func (s *TransactionService) GetAllTransactions(ctx context.Context) ([]*model.Transaction, error) {
	return s.txnRepo.GetAll(ctx)
}
