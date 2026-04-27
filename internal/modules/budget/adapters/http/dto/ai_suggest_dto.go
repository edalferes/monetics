package dto

import (
	"github.com/edalferes/monetics/internal/modules/budget/adapters/ai"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"
)

// SuggestCategoriesItem is a single row to be categorized by AI.
type SuggestCategoriesItem struct {
	Date        string  `json:"date" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Type        string  `json:"type" validate:"omitempty,oneof=expense income transfer"`
}

// SuggestCategoriesRequest is the payload for POST /transactions/import/ai-suggest.
type SuggestCategoriesRequest struct {
	AccountID uint                    `json:"account_id" validate:"required"`
	Items     []SuggestCategoriesItem `json:"items" validate:"required,min=1,dive"`
}

// SuggestCategoriesResponse is the response from ai-suggest.
type SuggestCategoriesResponse struct {
	Suggestions []ai.CategorySuggestion `json:"suggestions"`
	Usage       ai.Usage                `json:"usage"`
}

// ToSuggestCategoriesItems converts request items to use case items.
func ToSuggestCategoriesItems(in []SuggestCategoriesItem) []transaction.ImportItem {
	out := make([]transaction.ImportItem, len(in))
	for i, it := range in {
		out[i] = transaction.ImportItem{
			Date:        it.Date,
			Description: it.Description,
			Amount:      it.Amount,
			Type:        it.Type,
		}
	}
	return out
}

// ToSuggestCategoriesResponse converts use case output to response DTO.
func ToSuggestCategoriesResponse(out transaction.SuggestCategoriesOutput) SuggestCategoriesResponse {
	return SuggestCategoriesResponse{
		Suggestions: out.Suggestions,
		Usage:       out.Usage,
	}
}

// FeaturesResponse is returned by GET /config/features.
type FeaturesResponse struct {
	AIImport bool `json:"ai_import"`
}
