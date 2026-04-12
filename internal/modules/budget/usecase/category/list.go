package category

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// ListUseCase handles listing user categories
type ListUseCase struct {
	categoryRepo interfaces.CategoryRepository
	logger       logger.Logger
}

// NewListUseCase creates a new use case instance
func NewListUseCase(categoryRepo interfaces.CategoryRepository, log logger.Logger) *ListUseCase {
	return &ListUseCase{
		categoryRepo: categoryRepo,
		logger:       log.With().Str("usecase", "category.list").Logger(),
	}
}

// Execute lists all categories for a user, optionally filtered by type
func (uc *ListUseCase) Execute(ctx context.Context, userID uint, categoryType *domain.CategoryType) ([]domain.Category, error) {
	uc.logger.Debug().Uint("user_id", userID).Msg("listing categories")

	var categories []domain.Category
	var err error

	if categoryType != nil {
		categories, err = uc.categoryRepo.GetByType(ctx, userID, *categoryType)
	} else {
		categories, err = uc.categoryRepo.GetByUserID(ctx, userID)
	}
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to list categories")
		return nil, err
	}

	uc.logger.Info().Uint("user_id", userID).Int("count", len(categories)).Msg("categories listed successfully")
	return categories, nil
}
