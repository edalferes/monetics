package domain

import "time"

// AccountType represents the type of account
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"   // Conta corrente
	AccountTypeSavings  AccountType = "savings"    // Savings account
	AccountTypeCredit   AccountType = "credit_card" // Credit card
	AccountTypeCash     AccountType = "cash"       // Dinheiro
	AccountTypeInvest   AccountType = "investment" // Investimentos
)

// Account represents a financial account (bank account, credit card, cash, etc.)
//
// An Account belongs to a User and contains the balance and transaction history.
// It can represent different types of accounts like checking, savings, credit cards.
//
// Business rules:
//   - Each account must belong to a user
//   - Initial balance can be positive or negative (debt)
//   - Account name must be unique per user
//   - Balance is updated automatically with transactions
//
// Example:
//
//	account := &Account{
//		UserID:      1,
//		Name:        "Banco Inter",
//		Type:        AccountTypeChecking,
//		Balance:     5000.00,
//		Currency:    "BRL",
//		Description: "Conta corrente principal",
//	}
type Account struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	UserID      uint        `json:"user_id" gorm:"not null;index:idx_user_accounts;constraint:OnDelete:CASCADE"`
	Name        string      `json:"name" gorm:"not null;size:100"`
	Type        AccountType `json:"type" gorm:"not null;size:20"`
	Balance     float64     `json:"balance" gorm:"type:decimal(15,2);default:0"`
	Currency    string      `json:"currency" gorm:"size:3;default:'BRL'"`
	Description string      `json:"description,omitempty" gorm:"type:text"`
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func (Account) TableName() string {
	return "budget_accounts"
}
