package category

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// DeleteUseCase handles category deletion (soft delete)
type DeleteUseCase struct {
	categoryRepo    interfaces.CategoryRepository
	transactionRepo interfaces.TransactionRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

// NewDeleteUseCase creates a new use case instance
func NewDeleteUseCase(
	categoryRepo interfaces.CategoryRepository,
	transactionRepo interfaces.TransactionRepository,
	budgetRepo interfaces.BudgetRepository,
	log logger.Logger,
) *DeleteUseCase {
	return &DeleteUseCase{
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "category.delete").Logger(),
	}
}

// Execute soft-deletes a category
func (uc *DeleteUseCase) Execute(ctx context.Context, categoryID, userID uint) error {
	uc.logger.Debug().Uint("user_id", userID).Uint("category_id", categoryID).Msg("deleting category")

	if categoryID == 0 {
		uc.logger.Error().Msg("invalid category_id: cannot be zero")
		return errors.ErrCategoryNotFound
	}

	// Get existing category
	category, err := uc.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", categoryID).Msg("category not found")
		return err
	}

	// Ensure the category belongs to the user
	if category.UserID != userID {
		uc.logger.Warn().
			Uint("category_user_id", category.UserID).
			Uint("request_user_id", userID).
			Msg("unauthorized access: category belongs to different user")
		return errors.ErrCategoryNotFound
	}

	// Check if category is in use by transactions
	transactions, err := uc.transactionRepo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", categoryID).Msg("failed to check transactions")
		return errors.ErrInternalServer
	}
	if len(transactions) > 0 {
		uc.logger.Warn().Uint("category_id", categoryID).Int("transactions", len(transactions)).Msg("category in use by transactions")
		return errors.ErrCategoryInUse
	}

	// Check if category is in use by budgets
	budgets, err := uc.budgetRepo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", categoryID).Msg("failed to check budgets")
		return errors.ErrInternalServer
	}
	if len(budgets) > 0 {
		uc.logger.Warn().Uint("category_id", categoryID).Int("budgets", len(budgets)).Msg("category in use by budgets")
		return errors.ErrCategoryInUse
	}

	// Safe to delete
	err = uc.categoryRepo.Delete(ctx, categoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", categoryID).Msg("failed to delete category")
		return err
	}

	uc.logger.Info().Uint("category_id", categoryID).Msg("category deleted successfully")
	return nil
}
