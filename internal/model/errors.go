package model

import "fmt"

// TransferError represents different types of transfer errors
type TransferError string

const (
	// Same account transfer error
	ErrSameAccountTransfer TransferError = "cannot transfer to same account"

	// Amount validation errors
	ErrNegativeAmount TransferError = "transfer amount must be positive"
	ErrZeroAmount     TransferError = "transfer amount must be positive"

	// Account validation errors
	ErrInvalidAccountIDs     TransferError = "invalid account IDs"
	ErrSourceAccountNotFound TransferError = "source account not found"
	ErrDestAccountNotFound   TransferError = "destination account not found"

	// Balance errors
	ErrInsufficientFunds TransferError = "insufficient funds"

	// Transaction errors
	ErrFailedDebit     TransferError = "failed to debit source account"
	ErrFailedCredit    TransferError = "failed to credit destination account"
	ErrFailedRecordTxn TransferError = "failed to record transaction"

	// Service errors
	ErrServiceUnavailable TransferError = "service temporarily unavailable"
)

// Error returns the string representation of the error
func (e TransferError) Error() string {
	return string(e)
}

// AccountError represents different types of account errors
type AccountError string

const (
	ErrAccountIDRequired   AccountError = "account ID must be positive"
	ErrAccountNotFound     AccountError = "account not found"
	ErrAccountExists       AccountError = "account already exists"
	ErrNegativeBalance     AccountError = "account balance cannot be negative"
	ErrFailedCreateAccount AccountError = "failed to create account"
	ErrFailedGetAccount    AccountError = "failed to retrieve account"
)

// Error returns the string representation of the error
func (e AccountError) Error() string {
	return string(e)
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e TransferError) HTTPStatus() int {
	switch e {
	case ErrSameAccountTransfer, ErrNegativeAmount, ErrInvalidAccountIDs:
		return 400 // Bad Request
	case ErrSourceAccountNotFound, ErrDestAccountNotFound:
		return 404 // Not Found
	case ErrInsufficientFunds:
		return 422 // Unprocessable Entity
	case ErrFailedDebit, ErrFailedCredit, ErrFailedRecordTxn, ErrServiceUnavailable:
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e AccountError) HTTPStatus() int {
	switch e {
	case ErrAccountIDRequired, ErrNegativeBalance:
		return 400 // Bad Request
	case ErrAccountNotFound:
		return 404 // Not Found
	case ErrAccountExists:
		return 409 // Conflict
	case ErrFailedCreateAccount, ErrFailedGetAccount:
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}

// ErrorWithContext creates an error with additional context
func (e TransferError) WithContext(ctx string) error {
	return fmt.Errorf("%s: %s", string(e), ctx)
}

// ErrorWithContext creates an error with additional context
func (e AccountError) WithContext(ctx string) error {
	return fmt.Errorf("%s: %s", string(e), ctx)
}
