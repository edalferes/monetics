package dto

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=checking savings credit_card cash investment"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

// AccountResponse represents an account in API responses
type AccountResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Balance     float64   `json:"balance"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToAccountResponse converts domain.Account to AccountResponse
func ToAccountResponse(account domain.Account) AccountResponse {
	return AccountResponse{
		ID:          account.ID,
		UserID:      account.UserID,
		Name:        account.Name,
		Type:        string(account.Type),
		Balance:     account.Balance,
		Currency:    account.Currency,
		Description: account.Description,
		IsActive:    account.IsActive,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
}

// ToAccountResponseList converts []domain.Account to []AccountResponse
func ToAccountResponseList(accounts []domain.Account) []AccountResponse {
	responses := make([]AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = ToAccountResponse(account)
	}
	return responses
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	Type        *string `json:"type" validate:"omitempty,oneof=checking savings credit_card cash investment"`
	Currency    *string `json:"currency" validate:"omitempty,len=3"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

// AccountBalanceResponse represents an account with calculated balance details
type AccountBalanceResponse struct {
	Account        AccountResponse `json:"account"`
	CurrentBalance float64         `json:"current_balance"`
	TotalIncome    float64         `json:"total_income"`
	TotalExpense   float64         `json:"total_expense"`
	TotalTransfers float64         `json:"total_transfers"`
}
