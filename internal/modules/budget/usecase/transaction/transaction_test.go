package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edalferes/monetics/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	budgetErrors "github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"
)

func TestCreateUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	accountID := uint(10)
	categoryID := uint(20)
	now := time.Now()

	t.Run("should create transaction successfully", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		input := transaction.CreateInput{
			UserID:      userID,
			AccountID:   accountID,
			CategoryID:  categoryID,
			Amount:      100.50,
			Type:        domain.TransactionTypeExpense,
			Description: "Test transaction",
			Date:        now.Format(time.RFC3339),
		}

		accountObj := domain.Account{
			ID:        accountID,
			UserID:    userID,
			Name:      "Test Account",
			Type:      domain.AccountTypeChecking,
			Balance:   1000.00,
			Currency:  "USD",
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		categoryObj := domain.Category{
			ID:        categoryID,
			UserID:    userID,
			Name:      "Test Category",
			Type:      domain.CategoryTypeExpense,
			IsActive:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}

		expectedTx := domain.Transaction{
			UserID:      userID,
			AccountID:   accountID,
			CategoryID:  categoryID,
			Amount:      100.50,
			Type:        domain.TransactionTypeExpense,
			Description: "Test transaction",
			Status:      domain.TransactionStatusCompleted,
		}

		mockAccountRepo.On("GetByID", ctx, accountID).Return(accountObj, nil)
		mockCategoryRepo.On("GetByID", ctx, categoryID).Return(categoryObj, nil)
		mockTransactionRepo.On("Create", ctx, mock.MatchedBy(func(t domain.Transaction) bool {
			return t.UserID == userID &&
				t.AccountID == accountID &&
				t.CategoryID == categoryID &&
				t.Amount == 100.50 &&
				t.Type == domain.TransactionTypeExpense
		})).Return(expectedTx, nil)
		// Mock for async updateBudgetSpent goroutine
		mockBudgetRepo.On("GetByUserID", mock.Anything, userID).Return([]domain.Budget{}, nil).Maybe()

		result, err := usecase.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, accountID, result.AccountID)
		assert.Equal(t, categoryID, result.CategoryID)
		mockAccountRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should return error when account does not belong to user", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		input := transaction.CreateInput{
			UserID:     userID,
			AccountID:  999,
			CategoryID: categoryID,
			Amount:     100.50,
			Type:       domain.TransactionTypeExpense,
			Date:       now.Format(time.RFC3339),
		}

		mockAccountRepo.On("GetByID", ctx, uint(999)).Return(domain.Account{}, budgetErrors.ErrAccountNotFound)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrAccountNotFound, err)
		mockAccountRepo.AssertExpectations(t)
	})

	t.Run("should return error when category does not belong to user", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		input := transaction.CreateInput{
			UserID:     userID,
			AccountID:  accountID,
			CategoryID: 999,
			Amount:     100.50,
			Type:       domain.TransactionTypeExpense,
			Date:       now.Format(time.RFC3339),
		}

		accountObj := domain.Account{
			ID:       accountID,
			UserID:   userID,
			Name:     "Test Account",
			Type:     domain.AccountTypeChecking,
			IsActive: true,
		}

		mockAccountRepo.On("GetByID", ctx, accountID).Return(accountObj, nil)
		mockCategoryRepo.On("GetByID", ctx, uint(999)).Return(domain.Category{}, budgetErrors.ErrCategoryNotFound)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrCategoryNotFound, err)
		mockAccountRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid transaction type", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		input := transaction.CreateInput{
			UserID:     userID,
			AccountID:  accountID,
			CategoryID: categoryID,
			Amount:     100.50,
			Type:       "invalid",
			Date:       now.Format(time.RFC3339),
		}

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrInvalidTransactionType, err)
	})

	t.Run("should create transfer with destination account", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		destAccountID := uint(11)
		input := transaction.CreateInput{
			UserID:               userID,
			AccountID:            accountID,
			CategoryID:           categoryID,
			Amount:               100.50,
			Type:                 domain.TransactionTypeTransfer,
			Description:          "Transfer",
			Date:                 now.Format(time.RFC3339),
			DestinationAccountID: &destAccountID,
		}

		sourceAccount := domain.Account{
			ID:       accountID,
			UserID:   userID,
			Name:     "Source Account",
			IsActive: true,
		}

		destAccount := domain.Account{
			ID:       destAccountID,
			UserID:   userID,
			Name:     "Destination Account",
			IsActive: true,
		}

		categoryObj := domain.Category{
			ID:       categoryID,
			UserID:   userID,
			Name:     "Transfer Category",
			IsActive: true,
		}

		expectedTx := domain.Transaction{
			UserID:               userID,
			AccountID:            accountID,
			CategoryID:           categoryID,
			Amount:               100.50,
			Type:                 domain.TransactionTypeTransfer,
			Description:          "Transfer",
			Status:               domain.TransactionStatusCompleted,
			DestinationAccountID: &destAccountID,
		}

		mockAccountRepo.On("GetByID", ctx, accountID).Return(sourceAccount, nil)
		mockCategoryRepo.On("GetByID", ctx, categoryID).Return(categoryObj, nil)
		mockAccountRepo.On("GetByID", ctx, destAccountID).Return(destAccount, nil)
		mockTransactionRepo.On("CreateTransfer", ctx,
			mock.MatchedBy(func(t domain.Transaction) bool {
				return t.UserID == userID &&
					t.AccountID == accountID &&
					t.DestinationAccountID != nil &&
					*t.DestinationAccountID == destAccountID
			}),
			mock.MatchedBy(func(t domain.Transaction) bool {
				return t.UserID == userID &&
					t.AccountID == destAccountID &&
					t.DestinationAccountID != nil &&
					*t.DestinationAccountID == accountID
			}),
		).Return(expectedTx, expectedTx, nil)

		result, err := usecase.Execute(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result.DestinationAccountID)
		assert.Equal(t, destAccountID, *result.DestinationAccountID)
		mockAccountRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should return error when destination account not found", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		destAccountID := uint(999)
		input := transaction.CreateInput{
			UserID:               userID,
			AccountID:            accountID,
			CategoryID:           categoryID,
			Amount:               100.50,
			Type:                 domain.TransactionTypeTransfer,
			Date:                 now.Format(time.RFC3339),
			DestinationAccountID: &destAccountID,
		}

		sourceAccount := domain.Account{
			ID:       accountID,
			UserID:   userID,
			Name:     "Source Account",
			IsActive: true,
		}

		categoryObj := domain.Category{
			ID:       categoryID,
			UserID:   userID,
			Name:     "Transfer Category",
			IsActive: true,
		}

		mockAccountRepo.On("GetByID", ctx, accountID).Return(sourceAccount, nil)
		mockCategoryRepo.On("GetByID", ctx, categoryID).Return(categoryObj, nil)
		mockAccountRepo.On("GetByID", ctx, destAccountID).Return(domain.Account{}, budgetErrors.ErrAccountNotFound)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrAccountNotFound, err)
		mockAccountRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("should return error when destination account belongs to different user", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewCreateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		destAccountID := uint(11)
		input := transaction.CreateInput{
			UserID:               userID,
			AccountID:            accountID,
			CategoryID:           categoryID,
			Amount:               100.50,
			Type:                 domain.TransactionTypeTransfer,
			Date:                 now.Format(time.RFC3339),
			DestinationAccountID: &destAccountID,
		}

		sourceAccount := domain.Account{
			ID:       accountID,
			UserID:   userID,
			Name:     "Source Account",
			IsActive: true,
		}

		destAccount := domain.Account{
			ID:       destAccountID,
			UserID:   999, // Different user
			Name:     "Other User's Account",
			IsActive: true,
		}

		categoryObj := domain.Category{
			ID:       categoryID,
			UserID:   userID,
			Name:     "Transfer Category",
			IsActive: true,
		}

		mockAccountRepo.On("GetByID", ctx, accountID).Return(sourceAccount, nil)
		mockCategoryRepo.On("GetByID", ctx, categoryID).Return(categoryObj, nil)
		mockAccountRepo.On("GetByID", ctx, destAccountID).Return(destAccount, nil)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrUnauthorizedAccess, err)
		mockAccountRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
	})
}

func TestListUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	now := time.Now()

	t.Run("should list transactions successfully", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		usecase := transaction.NewListUseCase(mockRepo, logger.NewDefault())

		expectedTransactions := []domain.Transaction{
			{
				ID:          1,
				UserID:      userID,
				AccountID:   10,
				CategoryID:  20,
				Amount:      100.50,
				Type:        domain.TransactionTypeExpense,
				Description: "Transaction 1",
				Date:        now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          2,
				UserID:      userID,
				AccountID:   10,
				CategoryID:  21,
				Amount:      200.00,
				Type:        domain.TransactionTypeIncome,
				Description: "Transaction 2",
				Date:        now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		mockRepo.On("GetByUserIDPaginatedWithAllFilters", ctx, userID, 20, 0, (*domain.TransactionType)(nil), (*uint)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(expectedTransactions, nil)
		mockRepo.On("CountByUserIDWithAllFilters", ctx, userID, (*domain.TransactionType)(nil), (*uint)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(int64(2), nil)

		input := transaction.ListInput{
			UserID:   userID,
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Len(t, result.Transactions, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
		assert.Equal(t, 1, result.TotalPages)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error on repository failure", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		usecase := transaction.NewListUseCase(mockRepo, logger.NewDefault())

		expectedError := errors.New("database error")
		mockRepo.On("GetByUserIDPaginatedWithAllFilters", ctx, userID, 20, 0, (*domain.TransactionType)(nil), (*uint)(nil), (*uint)(nil), (*time.Time)(nil), (*time.Time)(nil)).Return(nil, expectedError)

		input := transaction.ListInput{
			UserID:   userID,
			Page:     1,
			PageSize: 20,
		}

		result, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Empty(t, result.Transactions)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetByIDUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	transactionID := uint(10)
	now := time.Now()

	t.Run("should get transaction by id successfully", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		usecase := transaction.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		expectedTransaction := domain.Transaction{
			ID:          transactionID,
			UserID:      userID,
			AccountID:   10,
			CategoryID:  20,
			Amount:      100.50,
			Type:        domain.TransactionTypeExpense,
			Description: "Test transaction",
			Date:        now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockRepo.On("GetByID", ctx, transactionID).Return(expectedTransaction, nil)

		result, err := usecase.Execute(ctx, userID, transactionID)

		assert.NoError(t, err)
		assert.Equal(t, expectedTransaction, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when transaction not found", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		usecase := transaction.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		mockRepo.On("GetByID", ctx, transactionID).Return(domain.Transaction{}, budgetErrors.ErrTransactionNotFound)

		result, err := usecase.Execute(ctx, userID, transactionID)

		assert.Error(t, err)
		assert.Equal(t, domain.Transaction{}, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	transactionID := uint(10)

	t.Run("should delete transaction successfully", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockBudgetRepo := new(MockBudgetRepository)
		usecase := transaction.NewDeleteUseCase(mockRepo, mockBudgetRepo, logger.NewDefault())

		existingTx := domain.Transaction{
			ID:        transactionID,
			UserID:    userID,
			AccountID: 10,
			Amount:    100.50,
		}

		mockRepo.On("GetByID", ctx, transactionID).Return(existingTx, nil)
		mockRepo.On("Delete", ctx, transactionID).Return(nil)
		mockBudgetRepo.On("GetByUserID", ctx, userID).Return([]domain.Budget{}, nil)

		err := usecase.Execute(ctx, userID, transactionID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})

	t.Run("should return error when transaction not found", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockBudgetRepo := new(MockBudgetRepository)
		usecase := transaction.NewDeleteUseCase(mockRepo, mockBudgetRepo, logger.NewDefault())

		mockRepo.On("GetByID", ctx, transactionID).Return(domain.Transaction{}, budgetErrors.ErrTransactionNotFound)

		err := usecase.Execute(ctx, userID, transactionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})

	t.Run("should return error when unauthorized", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)
		mockBudgetRepo := new(MockBudgetRepository)
		usecase := transaction.NewDeleteUseCase(mockRepo, mockBudgetRepo, logger.NewDefault())

		existingTx := domain.Transaction{
			ID:        transactionID,
			UserID:    999, // Different user
			AccountID: 10,
			Amount:    100.50,
		}

		mockRepo.On("GetByID", ctx, transactionID).Return(existingTx, nil)

		err := usecase.Execute(ctx, userID, transactionID)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrTransactionNotFound, err)
		mockRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})
}

func TestUpdateUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	userID := uint(1)
	transactionID := uint(10)
	accountID := uint(20)
	categoryID := uint(30)
	now := time.Now()

	t.Run("should update transaction successfully", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewUpdateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		existingTx := domain.Transaction{
			ID:          transactionID,
			UserID:      userID,
			AccountID:   accountID,
			CategoryID:  categoryID,
			Amount:      100.50,
			Type:        domain.TransactionTypeExpense,
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		newAmount := 200.00
		newDescription := "Updated description"

		input := transaction.UpdateInput{
			ID:          transactionID,
			UserID:      userID,
			Amount:      &newAmount,
			Description: &newDescription,
		}

		updatedTx := existingTx
		updatedTx.Amount = newAmount
		updatedTx.Description = newDescription

		mockTransactionRepo.On("GetByID", ctx, transactionID).Return(existingTx, nil)
		mockTransactionRepo.On("Update", ctx, mock.MatchedBy(func(t domain.Transaction) bool {
			return t.ID == transactionID &&
				t.Amount == newAmount &&
				t.Description == newDescription
		})).Return(updatedTx, nil)
		mockBudgetRepo.On("GetByUserID", ctx, userID).Return([]domain.Budget{}, nil)

		result, err := usecase.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, newAmount, result.Amount)
		assert.Equal(t, newDescription, result.Description)
		mockTransactionRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})

	t.Run("should return error when transaction not found", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewUpdateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		newAmount := 200.00

		input := transaction.UpdateInput{
			ID:     999,
			UserID: userID,
			Amount: &newAmount,
		}

		mockTransactionRepo.On("GetByID", ctx, uint(999)).Return(domain.Transaction{}, budgetErrors.ErrTransactionNotFound)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		mockTransactionRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})

	t.Run("should return error when unauthorized", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewUpdateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		existingTx := domain.Transaction{
			ID:        transactionID,
			UserID:    999, // Different user
			AccountID: accountID,
			Amount:    100.50,
		}

		newAmount := 200.00

		input := transaction.UpdateInput{
			ID:     transactionID,
			UserID: userID,
			Amount: &newAmount,
		}

		mockTransactionRepo.On("GetByID", ctx, transactionID).Return(existingTx, nil)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrTransactionNotFound, err)
		mockTransactionRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid amount", func(t *testing.T) {
		mockTransactionRepo := new(MockTransactionRepository)
		mockAccountRepo := new(MockAccountRepository)
		mockCategoryRepo := new(MockCategoryRepository)
		mockBudgetRepo := new(MockBudgetRepository)

		usecase := transaction.NewUpdateUseCase(mockTransactionRepo, mockAccountRepo, mockCategoryRepo, mockBudgetRepo, logger.NewDefault())

		existingTx := domain.Transaction{
			ID:        transactionID,
			UserID:    userID,
			AccountID: accountID,
			Amount:    100.50,
		}

		invalidAmount := -50.00

		input := transaction.UpdateInput{
			ID:     transactionID,
			UserID: userID,
			Amount: &invalidAmount,
		}

		mockTransactionRepo.On("GetByID", ctx, transactionID).Return(existingTx, nil)

		_, err := usecase.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, budgetErrors.ErrInvalidAmount, err)
		mockTransactionRepo.AssertExpectations(t)
		mockBudgetRepo.AssertExpectations(t)
	})
}
