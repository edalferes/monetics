package budget

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type ListUseCase struct {
	budgetRepo      interfaces.BudgetRepository
	transactionRepo interfaces.TransactionRepository
	logger          logger.Logger
}

func NewListUseCase(budgetRepo interfaces.BudgetRepository, transactionRepo interfaces.TransactionRepository, log logger.Logger) *ListUseCase {
	return &ListUseCase{
		budgetRepo:      budgetRepo,
		transactionRepo: transactionRepo,
		logger:          log.With().Str("usecase", "budget.list").Logger(),
	}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID uint) ([]domain.Budget, error) {
	uc.logger.Debug().Uint("user_id", userID).Msg("listing budgets")

	budgets, err := uc.budgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to get budgets")
		return nil, err
	}

	// Update spent for each budget atomically via SQL
	for i := range budgets {
		if err := uc.budgetRepo.UpdateSpentAtomic(ctx, budgets[i].ID, budgets[i].CategoryID, budgets[i].StartDate, budgets[i].EndDate); err != nil {
			uc.logger.Error().Err(err).Uint("budget_id", budgets[i].ID).Msg("failed to update spent atomically, skipping")
			continue
		}
		// Re-fetch the updated budget to get the new spent value
		updated, err := uc.budgetRepo.GetByID(ctx, budgets[i].ID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("budget_id", budgets[i].ID).Msg("failed to re-fetch budget after spent update")
			continue
		}
		budgets[i].Spent = updated.Spent
	}

	uc.logger.Info().Uint("user_id", userID).Int("count", len(budgets)).Msg("budgets listed successfully")
	return budgets, nil
}
