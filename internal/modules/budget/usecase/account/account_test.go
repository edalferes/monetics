package account_test

import (
	"context"
	"testing"

	"github.com/edalferes/monetics/pkg/logger"

	"github.com/stretchr/testify/assert"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/account"
)

func TestCreateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success with all fields", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewCreateUseCase(mockRepo, logger.NewDefault())

		input := account.CreateInput{
			UserID:      1,
			Name:        "My Checking Account",
			Type:        domain.AccountTypeChecking,
			Currency:    "USD",
			Description: "Main account",
		}

		expectedAccount := domain.Account{
			ID:          1,
			UserID:      1,
			Name:        "My Checking Account",
			Type:        domain.AccountTypeChecking,
			Currency:    "USD",
			Description: "Main account",
		}

		mockRepo.On("Create", ctx, domain.Account{
			UserID:      input.UserID,
			Name:        input.Name,
			Type:        input.Type,
			Currency:    input.Currency,
			Description: input.Description,
			IsActive:    true,
		}).Return(expectedAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, expectedAccount, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success with default currency", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewCreateUseCase(mockRepo, logger.NewDefault())

		input := account.CreateInput{
			UserID: 1,
			Name:   "My Savings",
			Type:   domain.AccountTypeSavings,
		}

		expectedAccount := domain.Account{
			ID:       1,
			UserID:   1,
			Name:     "My Savings",
			Type:     domain.AccountTypeSavings,
			Currency: "BRL",
		}

		mockRepo.On("Create", ctx, domain.Account{
			UserID:   input.UserID,
			Name:     input.Name,
			Type:     input.Type,
			Currency: "BRL",
			IsActive: true,
		}).Return(expectedAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, "BRL", result.Currency)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - empty name", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewCreateUseCase(mockRepo, logger.NewDefault())

		input := account.CreateInput{
			UserID: 1,
			Name:   "",
			Type:   domain.AccountTypeChecking,
		}

		result, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNameRequired, err)
		assert.Equal(t, domain.Account{}, result)
	})

	t.Run("error - invalid account type", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewCreateUseCase(mockRepo, logger.NewDefault())

		input := account.CreateInput{
			UserID: 1,
			Name:   "Test Account",
			Type:   "InvalidType",
		}

		result, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidAccountType, err)
		assert.Equal(t, domain.Account{}, result)
	})
}

func TestListUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success with accounts", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		mockTransactionRepo := new(MockTransactionRepository)
		uc := account.NewListUseCase(mockRepo, mockTransactionRepo, logger.NewDefault())

		expectedAccounts := []domain.Account{
			{ID: 1, UserID: 1, Name: "Account 1", Type: domain.AccountTypeChecking},
			{ID: 2, UserID: 1, Name: "Account 2", Type: domain.AccountTypeSavings},
		}

		mockRepo.On("GetByUserID", ctx, uint(1)).Return(expectedAccounts, nil)
		mockTransactionRepo.On("GetByAccountID", ctx, uint(1)).Return([]domain.Transaction{}, nil)
		mockTransactionRepo.On("GetByAccountID", ctx, uint(2)).Return([]domain.Transaction{}, nil)

		result, err := uc.Execute(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedAccounts, result)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("success with no accounts", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		mockTransactionRepo := new(MockTransactionRepository)
		uc := account.NewListUseCase(mockRepo, mockTransactionRepo, logger.NewDefault())

		mockRepo.On("GetByUserID", ctx, uint(1)).Return([]domain.Account{}, nil)

		result, err := uc.Execute(ctx, 1)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})
}

func TestGetByIDUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		expectedAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "Test Account",
			Type:   domain.AccountTypeChecking,
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(expectedAccount, nil)

		result, err := uc.Execute(ctx, 1, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedAccount, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - account not found", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrAccountNotFound)

		result, err := uc.Execute(ctx, 1, 999)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		assert.Equal(t, domain.Account{}, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - unauthorized access", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		accountFromDB := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "User 1 Account",
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(accountFromDB, nil)

		result, err := uc.Execute(ctx, 2, 1)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		assert.Equal(t, domain.Account{}, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - zero account ID", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewGetByIDUseCase(mockRepo, logger.NewDefault())

		result, err := uc.Execute(ctx, 1, 0)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		assert.Equal(t, domain.Account{}, result)
	})
}

func TestUpdateUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success - update name", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewUpdateUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:       1,
			UserID:   1,
			Name:     "Old Name",
			Type:     domain.AccountTypeChecking,
			Currency: "BRL",
		}

		newName := "New Name"
		input := account.UpdateInput{
			ID:     1,
			UserID: 1,
			Name:   &newName,
		}

		updatedAccount := existingAccount
		updatedAccount.Name = newName

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)
		mockRepo.On("Update", ctx, updatedAccount).Return(updatedAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, "New Name", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - update type", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewUpdateUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "Test Account",
			Type:   domain.AccountTypeChecking,
		}

		newType := domain.AccountTypeSavings
		input := account.UpdateInput{
			ID:     1,
			UserID: 1,
			Type:   &newType,
		}

		updatedAccount := existingAccount
		updatedAccount.Type = newType

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)
		mockRepo.On("Update", ctx, updatedAccount).Return(updatedAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, domain.AccountTypeSavings, result.Type)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - empty name", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewUpdateUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "Original Name",
		}

		emptyName := ""
		input := account.UpdateInput{
			ID:     1,
			UserID: 1,
			Name:   &emptyName,
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNameRequired, err)
		assert.Equal(t, domain.Account{}, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - invalid account type", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewUpdateUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Type:   domain.AccountTypeChecking,
		}

		invalidType := domain.AccountType("InvalidType")
		input := account.UpdateInput{
			ID:     1,
			UserID: 1,
			Type:   &invalidType,
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidAccountType, err)
		assert.Equal(t, domain.Account{}, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - unauthorized access", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewUpdateUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "User 1 Account",
		}

		newName := "New Name"
		input := account.UpdateInput{
			ID:     1,
			UserID: 2,
			Name:   &newName,
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)

		result, err := uc.Execute(ctx, input)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		assert.Equal(t, domain.Account{}, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewDeleteUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "Account to Delete",
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(nil)

		err := uc.Execute(ctx, 1, 1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - account not found", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewDeleteUseCase(mockRepo, logger.NewDefault())

		mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.ErrAccountNotFound)

		err := uc.Execute(ctx, 1, 999)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - unauthorized access", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewDeleteUseCase(mockRepo, logger.NewDefault())

		existingAccount := domain.Account{
			ID:     1,
			UserID: 1,
			Name:   "User 1 Account",
		}

		mockRepo.On("GetByID", ctx, uint(1)).Return(existingAccount, nil)

		err := uc.Execute(ctx, 2, 1)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error - zero account ID", func(t *testing.T) {
		mockRepo := new(MockAccountRepository)
		uc := account.NewDeleteUseCase(mockRepo, logger.NewDefault())

		err := uc.Execute(ctx, 1, 0)

		assert.Error(t, err)
		assert.Equal(t, errors.ErrAccountNotFound, err)
	})
}
