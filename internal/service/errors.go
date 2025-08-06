package service

import "errors"

var (
	ErrInvalidAccount  = errors.New("invalid account")
	ErrAccountNotFound = errors.New("account not found")
)
