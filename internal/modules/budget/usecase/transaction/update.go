package transaction

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// UpdateInput represents the input for updating a transaction
type UpdateInput struct {
	ID          uint
	UserID      uint
	AccountID   *uint
	CategoryID  *uint
	Type        *domain.TransactionType
	Amount      *float64
	Description *string
	Date        *string
}

// UpdateUseCase handles transaction updates
type UpdateUseCase struct {
	transactionRepo interfaces.TransactionRepository
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

// NewUpdateUseCase creates a new use case instance
func NewUpdateUseCase(
	transactionRepo interfaces.TransactionRepository,
	accountRepo interfaces.AccountRepository,
	categoryRepo interfaces.CategoryRepository,
	budgetRepo interfaces.BudgetRepository,
	log logger.Logger,
) *UpdateUseCase {
	return &UpdateUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "transaction.update").Logger(),
	}
}

// Execute updates an existing transaction
func (uc *UpdateUseCase) Execute(ctx context.Context, input UpdateInput) (domain.Transaction, error) {
	uc.logger.Debug().Uint("transaction_id", input.ID).Uint("user_id", input.UserID).Msg("updating transaction")

	if input.ID == 0 {
		uc.logger.Error().Msg("invalid transaction_id: cannot be zero")
		return domain.Transaction{}, errors.ErrTransactionNotFound
	}

	// Get existing transaction
	tx, err := uc.transactionRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", input.ID).Msg("transaction not found")
		return domain.Transaction{}, err
	}

	// Ensure the transaction belongs to the user
	if tx.UserID != input.UserID {
		uc.logger.Warn().
			Uint("transaction_user_id", tx.UserID).
			Uint("request_user_id", input.UserID).
			Msg("unauthorized access: transaction belongs to different user")
		return domain.Transaction{}, errors.ErrTransactionNotFound
	}

	// Validate and update account if provided
	if input.AccountID != nil {
		account, err := uc.accountRepo.GetByID(ctx, *input.AccountID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("account_id", *input.AccountID).Msg("account not found for update")
			return domain.Transaction{}, errors.ErrAccountNotFound
		}
		if account.UserID != input.UserID {
			uc.logger.Warn().Uint("account_user_id", account.UserID).Uint("request_user_id", input.UserID).Msg("unauthorized account access")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
		tx.AccountID = *input.AccountID
	}

	// Validate and update category if provided
	if input.CategoryID != nil {
		category, err := uc.categoryRepo.GetByID(ctx, *input.CategoryID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("category_id", *input.CategoryID).Msg("category not found for update")
			return domain.Transaction{}, errors.ErrCategoryNotFound
		}
		if category.UserID != input.UserID {
			uc.logger.Warn().Uint("category_user_id", category.UserID).Uint("request_user_id", input.UserID).Msg("unauthorized category access")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
		tx.CategoryID = *input.CategoryID
	}

	// Update type if provided
	if input.Type != nil {
		if !isValidTransactionType(*input.Type) {
			uc.logger.Error().Str("type", string(*input.Type)).Msg("invalid transaction type")
			return domain.Transaction{}, errors.ErrInvalidTransactionType
		}
		tx.Type = *input.Type
	}

	// Update amount if provided
	if input.Amount != nil {
		if *input.Amount <= 0 {
			uc.logger.Error().Msg("invalid amount: must be positive")
			return domain.Transaction{}, errors.ErrInvalidAmount
		}
		tx.Amount = *input.Amount
	}

	// Update description if provided
	if input.Description != nil {
		tx.Description = *input.Description
	}

	// Update date if provided
	if input.Date != nil {
		parsedDate, err := time.Parse(time.RFC3339, *input.Date)
		if err != nil {
			uc.logger.Error().Err(err).Str("date", *input.Date).Msg("invalid date format")
			return domain.Transaction{}, errors.ErrInvalidDate
		}
		tx.Date = parsedDate
		tx.Month = parsedDate.Format("2006-01")
		// Recompute status from the new date so it stays consistent with the
		// domain rule (future date = pending, otherwise completed).
		if tx.Status != domain.TransactionStatusCancelled {
			tx.Status = domain.ResolveStatus(parsedDate, time.Now())
		}
	}

	// Track categories that need budget recalculation
	oldCategoryID := tx.CategoryID
	affectedCategories := map[uint]bool{oldCategoryID: true}
	if input.CategoryID != nil && *input.CategoryID != oldCategoryID {
		affectedCategories[*input.CategoryID] = true
	}

	updatedTx, err := uc.transactionRepo.Update(ctx, tx)
	if err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", input.ID).Msg("failed to update transaction")
		return domain.Transaction{}, err
	}

	// Recalculate budget spent for affected categories
	for categoryID := range affectedCategories {
		uc.recalculateBudgetSpent(ctx, input.UserID, categoryID, updatedTx.Date)
	}

	uc.logger.Info().Uint("transaction_id", input.ID).Msg("transaction updated successfully")
	return updatedTx, nil
}

// recalculateBudgetSpent atomically recalculates spent for budgets matching the category
func (uc *UpdateUseCase) recalculateBudgetSpent(ctx context.Context, userID, categoryID uint, txDate time.Time) {
	budgets, err := uc.budgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to get budgets for spent recalculation")
		return
	}

	for _, b := range budgets {
		if b.CategoryID == categoryID && b.IsActive &&
			!txDate.Before(b.StartDate) && !txDate.After(b.EndDate) {
			if err := uc.budgetRepo.UpdateSpentAtomic(ctx, b.ID, categoryID, b.StartDate, b.EndDate); err != nil {
				uc.logger.Error().Err(err).Uint("budget_id", b.ID).Msg("failed to recalculate budget spent")
			}
		}
	}
}
