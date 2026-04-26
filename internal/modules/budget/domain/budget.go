package domain

import "time"

// BudgetPeriod represents the budget period type.
type BudgetPeriod string

const (
	BudgetPeriodDaily     BudgetPeriod = "daily"
	BudgetPeriodWeekly    BudgetPeriod = "weekly"
	BudgetPeriodMonthly   BudgetPeriod = "monthly"
	BudgetPeriodQuarterly BudgetPeriod = "quarterly"
	BudgetPeriodYearly    BudgetPeriod = "yearly"
	BudgetPeriodCustom    BudgetPeriod = "custom"
)

// Budget represents a budget plan for a category.
//
// Business rules:
//   - Each budget must belong to a user and category.
//   - Amount must be positive.
//   - Period dates must be valid (start before end).
//   - No overlapping budgets for the same (user, category).
//   - Spent amount is calculated from completed expense transactions.
type Budget struct {
	ID          uint
	UserID      uint
	CategoryID  uint
	Name        string
	Amount      float64
	Spent       float64
	Period      BudgetPeriod
	StartDate   time.Time
	EndDate     time.Time
	AlertAt     *float64
	IsActive    bool
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// RemainingAmount returns how much budget is left.
func (b *Budget) RemainingAmount() float64 {
	return b.Amount - b.Spent
}

// PercentageUsed returns the percentage of budget used.
func (b *Budget) PercentageUsed() float64 {
	if b.Amount == 0 {
		return 0
	}
	return (b.Spent / b.Amount) * 100
}

// IsOverBudget reports whether spending exceeded the budget amount.
func (b *Budget) IsOverBudget() bool {
	return b.Spent > b.Amount
}

// ShouldAlert reports whether the alert threshold is reached.
func (b *Budget) ShouldAlert() bool {
	if b.AlertAt == nil {
		return false
	}
	return b.PercentageUsed() >= *b.AlertAt
}

// Overlaps reports whether two budgets overlap in time for the same user/category.
func (b *Budget) Overlaps(other *Budget) bool {
	if b.UserID != other.UserID || b.CategoryID != other.CategoryID {
		return false
	}
	return b.StartDate.Before(other.EndDate) && other.StartDate.Before(b.EndDate)
}
