package errors

import "errors"

// Budget aggregate errors.
var (
	ErrBudgetNotFound      = errors.New("budget not found")
	ErrBudgetAlreadyExists = errors.New("budget already exists for this period")
	ErrInvalidBudgetPeriod = errors.New("invalid budget period")
	ErrInvalidBudgetDates  = errors.New("invalid budget dates")
	ErrBudgetOverlap       = errors.New("budget period overlaps with existing budget")
	ErrBudgetNameRequired  = errors.New("budget name is required")
	ErrInvalidBudgetAmount = errors.New("invalid budget amount")
	ErrInvalidDateRange    = errors.New("invalid date range: start date must be before end date")
)
