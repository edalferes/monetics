package errors

import "errors"

// Category aggregate errors.
var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryInUse         = errors.New("category is in use by transactions")
	ErrInvalidCategoryType   = errors.New("invalid category type")
	ErrCategoryNameRequired  = errors.New("category name is required")
	ErrInvalidCategoryID     = errors.New("invalid category ID")
)
