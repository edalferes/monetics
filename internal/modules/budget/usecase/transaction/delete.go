package transaction

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type DeleteUseCase struct {
	transactionRepo interfaces.TransactionRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

func NewDeleteUseCase(transactionRepo interfaces.TransactionRepository, budgetRepo interfaces.BudgetRepository, log logger.Logger) *DeleteUseCase {
	return &DeleteUseCase{
		transactionRepo: transactionRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "transaction.delete").Logger(),
	}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, userID, transactionID uint) error {
	uc.logger.Debug().Uint("user_id", userID).Uint("transaction_id", transactionID).Msg("deleting transaction")

	if transactionID == 0 {
		uc.logger.Error().Msg("invalid transaction_id: cannot be zero")
		return errors.ErrTransactionNotFound
	}

	tx, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", transactionID).Msg("transaction not found")
		return err
	}

	if tx.UserID != userID {
		uc.logger.Warn().
			Uint("transaction_user_id", tx.UserID).
			Uint("request_user_id", userID).
			Msg("unauthorized access: transaction belongs to different user")
		return errors.ErrTransactionNotFound
	}

	// Save transaction details before deleting for budget recalculation
	categoryID := tx.CategoryID
	txDate := tx.Date

	err = uc.transactionRepo.Delete(ctx, transactionID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", transactionID).Msg("failed to delete transaction")
		return err
	}

	// Recalculate budget spent for the affected category
	uc.recalculateBudgetSpent(ctx, userID, categoryID, txDate)

	uc.logger.Info().Uint("transaction_id", transactionID).Msg("transaction deleted successfully")
	return nil
}

// recalculateBudgetSpent atomically recalculates spent for budgets matching the category
func (uc *DeleteUseCase) recalculateBudgetSpent(ctx context.Context, userID, categoryID uint, txDate time.Time) {
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
