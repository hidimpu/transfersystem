package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hidimpu/transfersystem/internal/model"
	"github.com/hidimpu/transfersystem/internal/repository"
	"github.com/hidimpu/transfersystem/internal/utils"
)

type AccountService interface {
	CreateAccount(ctx context.Context, account *model.Account) error
	GetAccountByID(ctx context.Context, id int64) (*model.Account, error)
}

type accountService struct {
	accountRepo *repository.AccountRepository
	logger      *utils.Logger
}

func NewAccountService(repo *repository.AccountRepository) AccountService {
	return &accountService{
		accountRepo: repo,
		logger:      utils.GlobalLogger,
	}
}

func (s *accountService) CreateAccount(ctx context.Context, account *model.Account) error {
	// Log account creation attempt
	s.logger.LogAccount("CREATE_ATTEMPT", account.ID, account.Balance.String(), false)

	// Validate account data
	if account.ID <= 0 {
		s.logger.LogWarning("ACCOUNT_VALIDATION", fmt.Sprintf("Invalid account ID: %d", account.ID))
		return model.ErrAccountIDRequired
	}
	if account.Balance.IsNegative() {
		s.logger.LogWarning("ACCOUNT_VALIDATION", fmt.Sprintf("Negative balance: %s", account.Balance.String()))
		return model.ErrNegativeBalance
	}

	err := s.accountRepo.Create(ctx, account)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint" {
			s.logger.LogWarning("ACCOUNT_CREATE", fmt.Sprintf("Account %d already exists", account.ID))
			return model.ErrAccountExists.WithContext(fmt.Sprintf("account with ID %d already exists", account.ID))
		}
		s.logger.LogError("ACCOUNT_CREATE", "CREATE_ERROR", fmt.Sprintf("Failed to create account %d", account.ID), err)
		return model.ErrFailedCreateAccount
	}

	// Log successful account creation
	s.logger.LogAccount("CREATE_SUCCESS", account.ID, account.Balance.String(), true)
	return nil
}

func (s *accountService) GetAccountByID(ctx context.Context, id int64) (*model.Account, error) {
	// Log account retrieval attempt
	s.logger.LogAccount("GET_ATTEMPT", id, "N/A", false)

	// Validate account ID
	if id <= 0 {
		s.logger.LogWarning("ACCOUNT_VALIDATION", fmt.Sprintf("Invalid account ID: %d", id))
		return nil, model.ErrAccountIDRequired
	}

	account, err := s.accountRepo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			s.logger.LogWarning("ACCOUNT_GET", fmt.Sprintf("Account not found: %d", id))
			return nil, model.ErrAccountNotFound
		}
		s.logger.LogError("ACCOUNT_GET", "GET_ERROR", fmt.Sprintf("Failed to retrieve account %d", id), err)
		return nil, model.ErrFailedGetAccount
	}

	// Log successful account retrieval
	s.logger.LogAccount("GET_SUCCESS", id, account.Balance.String(), true)
	return account, nil
}
