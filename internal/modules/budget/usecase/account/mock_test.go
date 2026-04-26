package account_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// MockAccountRepository is a mock implementation of AccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id uint) (domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return domain.Account{}, args.Error(1)
	}
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(ctx context.Context, account domain.Account) (domain.Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAccountRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) CreateTransfer(ctx context.Context, debit, credit domain.Transaction) (domain.Transaction, domain.Transaction, error) {
	args := m.Called(ctx, debit, credit)
	return args.Get(0).(domain.Transaction), args.Get(1).(domain.Transaction), args.Error(2)
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uint) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return domain.Transaction{}, args.Error(1)
	}
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByUserIDPaginated(ctx context.Context, userID uint, limit, offset int) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByUserIDPaginatedWithFilters(ctx context.Context, userID uint, limit, offset int, startDate, endDate *time.Time) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, limit, offset, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByUserIDPaginatedWithAllFilters(ctx context.Context, userID uint, limit, offset int, txType *domain.TransactionType, accountID, categoryID *uint, startDate, endDate *time.Time) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, limit, offset, txType, accountID, categoryID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) CountByUserIDWithFilters(ctx context.Context, userID uint, startDate, endDate *time.Time) (int64, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) CountByUserIDWithAllFilters(ctx context.Context, userID uint, txType *domain.TransactionType, accountID, categoryID *uint, startDate, endDate *time.Time) (int64, error) {
	args := m.Called(ctx, userID, txType, accountID, categoryID, startDate, endDate)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) GetByAccountID(ctx context.Context, accountID uint) ([]domain.Transaction, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Transaction, error) {
	args := m.Called(ctx, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByType(ctx context.Context, userID uint, transactionType domain.TransactionType) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, transactionType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransactionRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
