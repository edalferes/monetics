package dto

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	AccountID            uint       `json:"account_id" validate:"required"`
	CategoryID           uint       `json:"category_id" validate:"required"`
	Type                 string     `json:"type" validate:"required,oneof=income expense transfer"`
	Amount               float64    `json:"amount" validate:"required,gt=0"`
	Description          string     `json:"description"`
	Date                 time.Time  `json:"date" validate:"required"`
	Tags                 []string   `json:"tags"`
	Attachments          []string   `json:"attachments"`
	IsRecurring          bool       `json:"is_recurring"`
	RecurrenceRule       string     `json:"recurrence_rule"`
	RecurrenceEnd        *time.Time `json:"recurrence_end"`
	DestinationAccountID *uint      `json:"destination_account_id"`
	TransferFee          *float64   `json:"transfer_fee"`
}

// AccountSummary represents account info in transaction response
type AccountSummary struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

// CategorySummary represents category info in transaction response
type CategorySummary struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID                   uint             `json:"id"`
	UserID               uint             `json:"user_id"`
	AccountID            uint             `json:"account_id"`
	Account              *AccountSummary  `json:"account,omitempty"`
	CategoryID           uint             `json:"category_id"`
	Category             *CategorySummary `json:"category,omitempty"`
	Type                 string           `json:"type"`
	Amount               float64          `json:"amount"`
	Description          string           `json:"description"`
	Date                 time.Time        `json:"date"`
	Status               string           `json:"status"`
	Tags                 []string         `json:"tags,omitempty"`
	Attachments          []string         `json:"attachments,omitempty"`
	IsRecurring          bool             `json:"is_recurring"`
	RecurrenceRule       string           `json:"recurrence_rule,omitempty"`
	RecurrenceEnd        *time.Time       `json:"recurrence_end,omitempty"`
	ParentID             *uint            `json:"parent_id,omitempty"`
	DestinationAccountID *uint            `json:"destination_account_id,omitempty"`
	TransferFee          *float64         `json:"transfer_fee,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

// ToTransactionResponse converts domain.Transaction to TransactionResponse
func ToTransactionResponse(transaction domain.Transaction) TransactionResponse {
	var account *AccountSummary
	if transaction.Account != nil {
		account = &AccountSummary{
			ID:       transaction.Account.ID,
			Name:     transaction.Account.Name,
			Type:     string(transaction.Account.Type),
			Currency: transaction.Account.Currency,
		}
	}

	var category *CategorySummary
	if transaction.Category != nil {
		category = &CategorySummary{
			ID:    transaction.Category.ID,
			Name:  transaction.Category.Name,
			Type:  string(transaction.Category.Type),
			Icon:  transaction.Category.Icon,
			Color: transaction.Category.Color,
		}
	}

	return TransactionResponse{
		ID:                   transaction.ID,
		UserID:               transaction.UserID,
		AccountID:            transaction.AccountID,
		Account:              account,
		CategoryID:           transaction.CategoryID,
		Category:             category,
		Type:                 string(transaction.Type),
		Amount:               transaction.Amount,
		Description:          transaction.Description,
		Date:                 transaction.Date,
		Status:               string(transaction.Status),
		Tags:                 transaction.Tags,
		Attachments:          transaction.Attachments,
		IsRecurring:          transaction.IsRecurring,
		RecurrenceRule:       transaction.RecurrenceRule,
		RecurrenceEnd:        transaction.RecurrenceEnd,
		ParentID:             transaction.ParentID,
		DestinationAccountID: transaction.DestinationAccountID,
		TransferFee:          transaction.TransferFee,
		CreatedAt:            transaction.CreatedAt,
		UpdatedAt:            transaction.UpdatedAt,
	}
}

// ToTransactionResponseList converts []domain.Transaction to []TransactionResponse
func ToTransactionResponseList(transactions []domain.Transaction) []TransactionResponse {
	responses := make([]TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = ToTransactionResponse(transaction)
	}
	return responses
}

// TransactionListResponse represents paginated transaction list response
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   PaginationMeta        `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// TransactionListOutput represents the use case output (avoiding circular import)
type TransactionListOutput struct {
	Transactions []domain.Transaction
	Total        int64
	Page         int
	PageSize     int
	TotalPages   int
}

// ToTransactionListResponse converts use case output to paginated response
func ToTransactionListResponse(result TransactionListOutput) TransactionListResponse {
	return TransactionListResponse{
		Transactions: ToTransactionResponseList(result.Transactions),
		Pagination: PaginationMeta{
			Total:      result.Total,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalPages: result.TotalPages,
		},
	}
}

// UpdateTransactionRequest represents the request to update a transaction
type UpdateTransactionRequest struct {
	AccountID            *uint      `json:"account_id"`
	CategoryID           *uint      `json:"category_id"`
	Type                 *string    `json:"type" validate:"omitempty,oneof=income expense transfer"`
	Amount               *float64   `json:"amount" validate:"omitempty,gt=0"`
	Description          *string    `json:"description" validate:"omitempty,max=500"`
	Date                 *time.Time `json:"date"`
	DestinationAccountID *uint      `json:"destination_account_id"`
}
