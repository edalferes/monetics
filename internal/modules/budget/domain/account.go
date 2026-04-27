// Package domain contains the core business entities for the budget module.
// All structs here are framework-agnostic: no GORM, JSON, or HTTP tags.
// JSON serialization is handled by the HTTP DTO layer; persistence is handled
// by the GORM model layer (adapters/repository/model).
package domain

import "time"

// AccountType represents the type of account.
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCredit   AccountType = "credit_card"
	AccountTypeCash     AccountType = "cash"
	AccountTypeInvest   AccountType = "investment"
)

// Account represents a financial account (bank account, credit card, cash, etc.).
//
// Business rules:
//   - Each account must belong to a user.
//   - Initial balance can be positive or negative (debt).
//   - Account name must be unique per user.
//   - Balance stored in this entity represents the opening/initial balance only.
//     The current balance MUST be derived from transactions via the
//     GetAccountBalanceUseCase (RN-A1). Do not mutate Balance from transaction flows.
type Account struct {
	ID     uint
	UserID uint
	Name   string
	Type   AccountType
	// Balance is the opening/initial balance set at account creation.
	// It is NOT updated when transactions are created. The current balance is
	// computed as: Balance + sum(income) - sum(expense) + net(transfers).
	Balance     float64
	Currency    string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
