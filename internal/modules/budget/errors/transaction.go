package errors

import "errors"

// Transaction aggregate errors.
var (
	ErrTransactionNotFound            = errors.New("transaction not found")
	ErrInvalidTransactionType         = errors.New("invalid transaction type")
	ErrInvalidTransactionStatus       = errors.New("invalid transaction status")
	ErrInvalidAmount                  = errors.New("invalid amount")
	ErrInvalidDate                    = errors.New("invalid date")
	ErrTransferSameAccount            = errors.New("cannot transfer to same account")
	ErrTransactionCancelled           = errors.New("transaction is already cancelled")
	ErrTransactionDescriptionRequired = errors.New("transaction description is required")
	ErrTransferRequiresDestination    = errors.New("transfer requires destination account")
)
