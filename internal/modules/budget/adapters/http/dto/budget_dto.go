package dto

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// CreateBudgetRequest represents the request to create a budget
type CreateBudgetRequest struct {
	CategoryID  uint      `json:"category_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Period      string    `json:"period" validate:"required,oneof=daily weekly monthly quarterly yearly custom"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	AlertAt     *float64  `json:"alert_at"`
	Description string    `json:"description"`
}

// BudgetResponse represents a budget in API responses
type BudgetResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	CategoryID     uint      `json:"category_id"`
	Name           string    `json:"name"`
	Amount         float64   `json:"amount"`
	Spent          float64   `json:"spent"`
	Remaining      float64   `json:"remaining"`
	PercentageUsed float64   `json:"percentage_used"`
	Period         string    `json:"period"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	AlertAt        *float64  `json:"alert_at,omitempty"`
	IsActive       bool      `json:"is_active"`
	IsOverBudget   bool      `json:"is_over_budget"`
	ShouldAlert    bool      `json:"should_alert"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ToBudgetResponse converts domain.Budget to BudgetResponse
func ToBudgetResponse(budget domain.Budget) BudgetResponse {
	return BudgetResponse{
		ID:             budget.ID,
		UserID:         budget.UserID,
		CategoryID:     budget.CategoryID,
		Name:           budget.Name,
		Amount:         budget.Amount,
		Spent:          budget.Spent,
		Remaining:      budget.RemainingAmount(),
		PercentageUsed: budget.PercentageUsed(),
		Period:         string(budget.Period),
		StartDate:      budget.StartDate,
		EndDate:        budget.EndDate,
		AlertAt:        budget.AlertAt,
		IsActive:       budget.IsActive,
		IsOverBudget:   budget.IsOverBudget(),
		ShouldAlert:    budget.ShouldAlert(),
		Description:    budget.Description,
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      budget.UpdatedAt,
	}
}

// ToBudgetResponseList converts []domain.Budget to []BudgetResponse
func ToBudgetResponseList(budgets []domain.Budget) []BudgetResponse {
	responses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		responses[i] = ToBudgetResponse(budget)
	}
	return responses
}

// UpdateBudgetRequest represents the request to update a budget
type UpdateBudgetRequest struct {
	Name        *string    `json:"name" validate:"omitempty,min=3,max=100"`
	Amount      *float64   `json:"amount" validate:"omitempty,gt=0"`
	Period      *string    `json:"period" validate:"omitempty,oneof=daily weekly monthly quarterly yearly custom"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	AlertAt     *float64   `json:"alert_at" validate:"omitempty,gte=0,lte=100"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
}
