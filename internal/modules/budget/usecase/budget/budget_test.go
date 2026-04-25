package budget_test

import (
	"context"
	"testing"
	"time"

	"github.com/edalferes/monetics/pkg/logger"

	"github.com/stretchr/testify/assert"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/budget"
)

func TestCreateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockBudgetRepo := new(MockBudgetRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		uc := budget.NewCreateUseCase(mockBudgetRepo, mockCategoryRepo, logger.NewDefault())

		startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

		input := budget.CreateInput{
			UserID:     1,
			CategoryID: 1,
			Name:       "Janeiro 2025",
			Amount:     2000.00,
			Period:     domain.BudgetPeriodMonthly,
			StartDate:  startDate,
			EndDate:    endDate,
		}

		category := domain.Category{ID: 1, UserID: 1, Name: "Alimentação"}
		mockCategoryRepo.On("GetByID", ctx, uint(1)).Return(category, nil)
		mockBudgetRepo.On("GetOverlapping", ctx, uint(1), uint(1), startDate, endDate, (*uint)(nil)).Return([]domain.Budget{}, nil)

		expectedBudget := domain.Budget{
			ID:         1,
			UserID:     1,
			CategoryID: 1,
			Name:       "Janeiro 2025",
			Amount:     2000.00,
			Period:     domain.BudgetPeriodMonthly,
			StartDate:  startDate,
			EndDate:    endDate,
			IsActive:   true,
			Spent:      0,
		}

		mockBudgetRepo.On("Create", ctx, domain.Budget{
			UserID:     1,
			CategoryID: 1,
			Name:       "Janeiro 2025",
			Amount:     2000.00,
			Period:     domain.BudgetPeriodMonthly,
			StartDate:  startDate,
			EndDate:    endDate,
			IsActive:   true,
			Spent:      0,
		}).Return(expectedBudget, nil)

		result, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, expectedBudget.ID, result.ID)
		mockBudgetRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("error - invalid amount", func(t *testing.T) {
		mockBudgetRepo := new(MockBudgetRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		uc := budget.NewCreateUseCase(mockBudgetRepo, mockCategoryRepo, logger.NewDefault())

		input := budget.CreateInput{
			UserID:     1,
			CategoryID: 1,
			Name:       "Test",
			Amount:     -100,
			Period:     domain.BudgetPeriodMonthly,
		}

		_, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidBudgetAmount, err)
	})
}

func TestListUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBudgetRepository)
		mockTransactionRepo := new(MockTransactionRepository)
		uc := budget.NewListUseCase(mockRepo, mockTransactionRepo, logger.NewDefault())

		startDate1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate1 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
		startDate2 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		endDate2 := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)

		expected := []domain.Budget{
			{ID: 1, UserID: 1, Name: "Janeiro", CategoryID: 1, Amount: 1000.0, StartDate: startDate1, EndDate: endDate1},
			{ID: 2, UserID: 1, Name: "Fevereiro", CategoryID: 2, Amount: 1500.0, StartDate: startDate2, EndDate: endDate2},
		}

		mockRepo.On("GetByUserID", ctx, uint(1)).Return(expected, nil)
		mockRepo.On("UpdateSpentAtomic", ctx, uint(1), uint(1), startDate1, endDate1).Return(nil)
		mockRepo.On("GetByID", ctx, uint(1)).Return(expected[0], nil)
		mockRepo.On("UpdateSpentAtomic", ctx, uint(2), uint(2), startDate2, endDate2).Return(nil)
		mockRepo.On("GetByID", ctx, uint(2)).Return(expected[1], nil)

		result, err := uc.Execute(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})
}

func TestGetByIDUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBudgetRepository)
		uc := budget.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		expected := domain.Budget{ID: 1, UserID: 1, Name: "Janeiro"}
		mockRepo.On("GetByID", ctx, uint(1)).Return(expected, nil)

		result, err := uc.Execute(ctx, 1, 1)

		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - unauthorized", func(t *testing.T) {
		mockRepo := new(MockBudgetRepository)
		uc := budget.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		budgetFromDB := domain.Budget{ID: 1, UserID: 2}
		mockRepo.On("GetByID", ctx, uint(1)).Return(budgetFromDB, nil)

		_, err := uc.Execute(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrBudgetNotFound, err)
	})
}

func TestDeleteUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBudgetRepository)
		uc := budget.NewDeleteUseCase(mockRepo, logger.NewDefault())

		budgetFromDB := domain.Budget{ID: 1, UserID: 1}
		mockRepo.On("GetByID", ctx, uint(1)).Return(budgetFromDB, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := uc.Execute(ctx, 1, 1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
