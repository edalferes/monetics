package domain

import "time"

// TransactionType represents the type of transaction.
type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// Transaction represents a financial transaction (income, expense, or transfer).
//
// Business rules:
//   - Each transaction must belong to a user.
//   - Must have an account (source for expenses/income).
//   - Transfers must have source and destination accounts.
//   - Amount must be positive.
//   - Date can be future for planned transactions (status pending).
//   - Completed transactions update account balance.
type Transaction struct {
	ID          uint
	UserID      uint
	AccountID   uint
	CategoryID  uint
	Type        TransactionType
	Amount      float64
	Description string
	Date        time.Time
	Month       string
	Status      TransactionStatus

	// Transfer-specific fields.
	DestinationAccountID *uint
	TransferFee          *float64

	// Recurrence.
	IsRecurring    bool
	RecurrenceRule string
	RecurrenceEnd  *time.Time
	ParentID       *uint

	// Metadata.
	Tags        []string
	Attachments []string

	CreatedAt time.Time
	UpdatedAt time.Time

	// Optional relations populated by the repository.
	Account  *Account
	Category *Category
}

// ResolveStatus picks the right status given the transaction date and a
// reference "now". Future-dated entries cannot be completed and are coerced
// to pending, while past or current dates default to completed.
func ResolveStatus(date, now time.Time) TransactionStatus {
	if date.After(now) {
		return TransactionStatusPending
	}
	return TransactionStatusCompleted
}
