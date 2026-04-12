package budget

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type CreateUseCase struct {
	budgetRepo   interfaces.BudgetRepository
	categoryRepo interfaces.CategoryRepository
	logger       logger.Logger
}

func NewCreateUseCase(budgetRepo interfaces.BudgetRepository, categoryRepo interfaces.CategoryRepository, log logger.Logger) *CreateUseCase {
	return &CreateUseCase{
		budgetRepo:   budgetRepo,
		categoryRepo: categoryRepo,
		logger:       log.With().Str("usecase", "budget.create").Logger(),
	}
}

type CreateInput struct {
	UserID      uint
	CategoryID  uint
	Name        string
	Amount      float64
	Period      domain.BudgetPeriod
	StartDate   time.Time
	EndDate     time.Time
	AlertAt     *float64
	Description string
}

func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (domain.Budget, error) {
	uc.logger.Debug().
		Uint("user_id", input.UserID).
		Uint("category_id", input.CategoryID).
		Str("name", input.Name).
		Str("period", string(input.Period)).
		Msg("creating budget")

	if input.UserID == 0 {
		uc.logger.Error().Msg("invalid user_id: cannot be zero")
		return domain.Budget{}, errors.ErrInvalidUserID
	}
	if input.Name == "" {
		uc.logger.Error().Msg("budget name is required")
		return domain.Budget{}, errors.ErrBudgetNameRequired
	}
	if input.Amount <= 0 {
		uc.logger.Error().Msg("invalid budget amount: must be positive")
		return domain.Budget{}, errors.ErrInvalidBudgetAmount
	}
	if input.CategoryID == 0 {
		uc.logger.Error().Msg("invalid category_id: cannot be zero")
		return domain.Budget{}, errors.ErrInvalidCategoryID
	}
	if !isValidBudgetPeriod(input.Period) {
		uc.logger.Error().Str("period", string(input.Period)).Msg("invalid budget period")
		return domain.Budget{}, errors.ErrInvalidBudgetPeriod
	}
	if input.StartDate.After(input.EndDate) || input.StartDate.Equal(input.EndDate) {
		uc.logger.Error().Msg("invalid date range: start_date must be before end_date")
		return domain.Budget{}, errors.ErrInvalidDateRange
	}

	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", input.CategoryID).Msg("category not found")
		return domain.Budget{}, errors.ErrCategoryNotFound
	}
	if category.UserID != input.UserID {
		uc.logger.Warn().
			Uint("category_user_id", category.UserID).
			Uint("request_user_id", input.UserID).
			Msg("unauthorized access: category belongs to different user")
		return domain.Budget{}, errors.ErrUnauthorizedAccess
	}

	// Check for overlapping budgets in the same category and period
	overlapping, err := uc.budgetRepo.GetOverlapping(ctx, input.UserID, input.CategoryID, input.StartDate, input.EndDate, nil)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to check for overlapping budgets")
		return domain.Budget{}, err
	}
	if len(overlapping) > 0 {
		uc.logger.Warn().
			Uint("category_id", input.CategoryID).
			Int("overlapping_count", len(overlapping)).
			Msg("budget period overlaps with existing budget")
		return domain.Budget{}, errors.ErrBudgetOverlap
	}

	budgetEntity := domain.Budget{
		UserID:      input.UserID,
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		Amount:      input.Amount,
		Period:      input.Period,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		AlertAt:     input.AlertAt,
		Description: input.Description,
		IsActive:    true,
		Spent:       0,
	}

	createdBudget, err := uc.budgetRepo.Create(ctx, budgetEntity)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", input.UserID).Msg("failed to create budget")
		return domain.Budget{}, err
	}

	uc.logger.Info().
		Uint("budget_id", createdBudget.ID).
		Uint("user_id", createdBudget.UserID).
		Str("name", createdBudget.Name).
		Msg("budget created successfully")

	return createdBudget, nil
}
