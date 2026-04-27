package domain

import "time"

// CategoryType represents whether a category is for income or expense.
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a transaction category for budget organization.
//
// Business rules:
//   - Each category must belong to a user.
//   - Category name must be unique per user and type.
//   - Categories may have an associated icon/color for UI.
type Category struct {
	ID          uint
	UserID      uint
	Name        string
	Type        CategoryType
	Icon        string
	Color       string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
