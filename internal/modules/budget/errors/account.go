package errors

import "errors"

// Account aggregate errors.
var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidAccountType   = errors.New("invalid account type")
	ErrAccountNameRequired  = errors.New("account name is required")
	ErrInvalidAccountID     = errors.New("invalid account ID")
)
