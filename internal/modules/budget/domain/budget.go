package domain

import "time"

// BudgetPeriod represents the budget period type
type BudgetPeriod string

const (
	BudgetPeriodDaily     BudgetPeriod = "daily"     // Diário
	BudgetPeriodWeekly    BudgetPeriod = "weekly"    // Semanal
	BudgetPeriodMonthly   BudgetPeriod = "monthly"   // Mensal
	BudgetPeriodQuarterly BudgetPeriod = "quarterly" // Trimestral
	BudgetPeriodYearly    BudgetPeriod = "yearly"    // Annual
	BudgetPeriodCustom    BudgetPeriod = "custom"    // Personalizado
)

// Budget represents a budget plan for a category
//
// Budgets help users set spending limits for categories over time periods.
// They enable tracking of planned vs actual spending.
//
// Business rules:
//   - Each budget must belong to a user and category
//   - Amount must be positive
//   - Period dates must be valid (start before end)
//   - Cannot have overlapping budgets for same category
//   - Spent amount is calculated from transactions
//
// Example:
//
//	budget := &Budget{
//		UserID:     1,
//		CategoryID: 5,
//		Name:       "Orçamento Alimentação Janeiro",
//		Amount:     2000.00,
//		Period:     BudgetPeriodMonthly,
//		StartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
//		EndDate:    time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC),
//	}
type Budget struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	UserID      uint         `json:"user_id" gorm:"not null;index:idx_user_budgets;constraint:OnDelete:CASCADE"`
	CategoryID  uint         `json:"category_id" gorm:"not null;index:idx_category_budgets"`
	Name        string       `json:"name" gorm:"not null;size:200"`
	Amount      float64      `json:"amount" gorm:"type:decimal(15,2);not null"`
	Spent       float64      `json:"spent" gorm:"type:decimal(15,2);default:0"`
	Period      BudgetPeriod `json:"period" gorm:"not null;size:20"`
	StartDate   time.Time    `json:"start_date" gorm:"not null;index:idx_budget_period"`
	EndDate     time.Time    `json:"end_date" gorm:"not null;index:idx_budget_period"`
	AlertAt     *float64     `json:"alert_at,omitempty" gorm:"type:decimal(5,2)"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	Description string       `json:"description,omitempty" gorm:"type:text"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (Budget) TableName() string {
	return "budget_budgets"
}

// RemainingAmount returns how much budget is left
func (b *Budget) RemainingAmount() float64 {
	return b.Amount - b.Spent
}

// PercentageUsed returns the percentage of budget used
func (b *Budget) PercentageUsed() float64 {
	if b.Amount == 0 {
		return 0
	}
	return (b.Spent / b.Amount) * 100
}

// IsOverBudget checks if spending exceeded the budget
func (b *Budget) IsOverBudget() bool {
	return b.Spent > b.Amount
}

// ShouldAlert checks if alert threshold is reached
func (b *Budget) ShouldAlert() bool {
	if b.AlertAt == nil {
		return false
	}
	return b.PercentageUsed() >= *b.AlertAt
}
