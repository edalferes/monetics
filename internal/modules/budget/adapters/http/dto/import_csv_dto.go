package dto

import "github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"

// ImportTransactionItem represents a single transaction item in the import request
type ImportTransactionItem struct {
	Date        string  `json:"date" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	CategoryID  uint    `json:"category_id" validate:"required"`
}

// ImportTransactionsRequest represents the request to bulk import transactions
type ImportTransactionsRequest struct {
	AccountID    uint                    `json:"account_id" validate:"required"`
	Transactions []ImportTransactionItem `json:"transactions" validate:"required,min=1,dive"`
}

// ImportTransactionsResponse represents the response for bulk import
type ImportTransactionsResponse struct {
	Imported int                        `json:"imported"`
	Errors   []transaction.ImportError  `json:"errors,omitempty"`
}

// ToImportTransactionsResponse converts use case result to response DTO
func ToImportTransactionsResponse(result transaction.ImportResult) ImportTransactionsResponse {
	return ImportTransactionsResponse{
		Imported: result.Imported,
		Errors:   result.Errors,
	}
}
