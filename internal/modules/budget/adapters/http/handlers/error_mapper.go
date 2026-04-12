package handlers

import (
	"errors"
	"net/http"

	budgetErrors "github.com/edalferes/monetics/internal/modules/budget/errors"
)

// errorToHTTPStatus maps domain errors to appropriate HTTP status codes
func errorToHTTPStatus(err error) int {
	switch {
	// 404 Not Found
	case errors.Is(err, budgetErrors.ErrAccountNotFound),
		errors.Is(err, budgetErrors.ErrCategoryNotFound),
		errors.Is(err, budgetErrors.ErrTransactionNotFound),
		errors.Is(err, budgetErrors.ErrBudgetNotFound):
		return http.StatusNotFound

	// 403 Forbidden
	case errors.Is(err, budgetErrors.ErrUnauthorizedAccess),
		errors.Is(err, budgetErrors.ErrUnauthorized):
		return http.StatusForbidden

	// 409 Conflict
	case errors.Is(err, budgetErrors.ErrAccountAlreadyExists),
		errors.Is(err, budgetErrors.ErrCategoryAlreadyExists),
		errors.Is(err, budgetErrors.ErrBudgetAlreadyExists),
		errors.Is(err, budgetErrors.ErrBudgetOverlap),
		errors.Is(err, budgetErrors.ErrCategoryInUse):
		return http.StatusConflict

	// 422 Unprocessable Entity
	case errors.Is(err, budgetErrors.ErrTransferSameAccount),
		errors.Is(err, budgetErrors.ErrTransferRequiresDestination),
		errors.Is(err, budgetErrors.ErrTransactionCancelled),
		errors.Is(err, budgetErrors.ErrInsufficientBalance):
		return http.StatusUnprocessableEntity

	// 400 Bad Request
	case errors.Is(err, budgetErrors.ErrInvalidAccountType),
		errors.Is(err, budgetErrors.ErrAccountNameRequired),
		errors.Is(err, budgetErrors.ErrInvalidAccountID),
		errors.Is(err, budgetErrors.ErrInvalidCategoryType),
		errors.Is(err, budgetErrors.ErrCategoryNameRequired),
		errors.Is(err, budgetErrors.ErrInvalidCategoryID),
		errors.Is(err, budgetErrors.ErrInvalidTransactionType),
		errors.Is(err, budgetErrors.ErrInvalidTransactionStatus),
		errors.Is(err, budgetErrors.ErrInvalidAmount),
		errors.Is(err, budgetErrors.ErrInvalidDate),
		errors.Is(err, budgetErrors.ErrTransactionDescriptionRequired),
		errors.Is(err, budgetErrors.ErrInvalidBudgetPeriod),
		errors.Is(err, budgetErrors.ErrInvalidBudgetDates),
		errors.Is(err, budgetErrors.ErrBudgetNameRequired),
		errors.Is(err, budgetErrors.ErrInvalidBudgetAmount),
		errors.Is(err, budgetErrors.ErrInvalidDateRange),
		errors.Is(err, budgetErrors.ErrInvalidUserID),
		errors.Is(err, budgetErrors.ErrInvalidData):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
