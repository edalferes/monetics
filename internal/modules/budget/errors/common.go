package errors

import "errors"

// Common errors shared across the budget bounded context.
var (
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrInvalidData    = errors.New("invalid data")
	ErrInternalServer = errors.New("internal server error")
	ErrInvalidUserID  = errors.New("invalid user ID")
)

// ErrUnauthorizedAccess is kept as an alias of ErrUnauthorized for backward compatibility
// during refactoring; prefer ErrUnauthorized in new code.
var ErrUnauthorizedAccess = ErrUnauthorized
