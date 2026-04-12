package transaction

import (
	"context"
	"fmt"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/helpers"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// ImportItem represents a single transaction to import
type ImportItem struct {
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	CategoryID  uint    `json:"category_id"`
}

// ImportInput represents the input for bulk import
type ImportInput struct {
	UserID       uint
	AccountID    uint
	Transactions []ImportItem
}

// ImportResult represents the result of the import
type ImportResult struct {
	Imported int           `json:"imported"`
	Errors   []ImportError `json:"errors,omitempty"`
}

// ImportError represents a single row error during import
type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportCSVUseCase handles importing transactions from structured data
type ImportCSVUseCase struct {
	transactionRepo interfaces.TransactionRepository
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

// NewImportCSVUseCase creates a new ImportCSVUseCase
func NewImportCSVUseCase(
	transactionRepo interfaces.TransactionRepository,
	accountRepo interfaces.AccountRepository,
	categoryRepo interfaces.CategoryRepository,
	budgetRepo interfaces.BudgetRepository,
	log logger.Logger,
) *ImportCSVUseCase {
	return &ImportCSVUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "transaction.import").Logger(),
	}
}

// Execute processes a list of transaction items and creates expense transactions
func (uc *ImportCSVUseCase) Execute(ctx context.Context, input ImportInput) (ImportResult, error) {
	uc.logger.Info().
		Uint("user_id", input.UserID).
		Uint("account_id", input.AccountID).
		Int("items", len(input.Transactions)).
		Msg("starting bulk import")

	if input.UserID == 0 {
		return ImportResult{}, errors.ErrInvalidUserID
	}

	if len(input.Transactions) == 0 {
		return ImportResult{}, errors.ErrInvalidData
	}

	// Validate account exists and belongs to user
	account, err := uc.accountRepo.GetByID(ctx, input.AccountID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("account_id", input.AccountID).Msg("account not found")
		return ImportResult{}, errors.ErrAccountNotFound
	}
	if account.UserID != input.UserID {
		return ImportResult{}, errors.ErrUnauthorizedAccess
	}

	// Pre-validate and cache all unique category IDs
	categoryCache := make(map[uint]bool)
	for _, item := range input.Transactions {
		if item.CategoryID == 0 {
			return ImportResult{}, fmt.Errorf("all transactions must have a category_id")
		}
		categoryCache[item.CategoryID] = false
	}
	for catID := range categoryCache {
		cat, err := uc.categoryRepo.GetByID(ctx, catID)
		if err != nil {
			return ImportResult{}, errors.ErrCategoryNotFound
		}
		if cat.UserID != input.UserID {
			return ImportResult{}, errors.ErrUnauthorizedAccess
		}
		categoryCache[catID] = true
	}

	result := ImportResult{}
	affectedCategories := make(map[uint]bool)

	for i, item := range input.Transactions {
		rowNum := i + 1

		if item.Amount <= 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Message: "amount must be positive",
			})
			continue
		}

		date, err := helpers.ParseFlexibleDate(item.Date)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Message: fmt.Sprintf("invalid date '%s'", item.Date),
			})
			continue
		}

		tx := domain.Transaction{
			UserID:      input.UserID,
			AccountID:   input.AccountID,
			CategoryID:  item.CategoryID,
			Type:        domain.TransactionTypeExpense,
			Amount:      item.Amount,
			Description: item.Description,
			Date:        date,
			Month:       date.Format("2006-01"),
			Status:      domain.TransactionStatusCompleted,
		}

		_, err = uc.transactionRepo.Create(ctx, tx)
		if err != nil {
			uc.logger.Error().Err(err).Int("row", rowNum).Msg("failed to create transaction")
			result.Errors = append(result.Errors, ImportError{
				Row:     rowNum,
				Message: fmt.Sprintf("failed to save: %s", err.Error()),
			})
			continue
		}

		result.Imported++
		affectedCategories[item.CategoryID] = true
	}

	// Update budgets for all affected categories
	if result.Imported > 0 {
		for catID := range affectedCategories {
			uc.updateBudgetsAfterImport(ctx, input.UserID, catID)
		}
	}

	uc.logger.Info().
		Int("imported", result.Imported).
		Int("errors", len(result.Errors)).
		Msg("bulk import completed")

	return result, nil
}

// updateBudgetsAfterImport recalculates budget spent for the category
func (uc *ImportCSVUseCase) updateBudgetsAfterImport(ctx context.Context, userID, categoryID uint) {
	budgets, err := uc.budgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to get budgets for post-import update")
		return
	}

	for _, budget := range budgets {
		if budget.CategoryID == categoryID && budget.IsActive {
			err = uc.budgetRepo.UpdateSpentAtomic(ctx, budget.ID, categoryID, budget.StartDate, budget.EndDate)
			if err != nil {
				uc.logger.Error().Err(err).Uint("budget_id", budget.ID).Msg("failed to update budget spent after import")
			}
		}
	}
}
