package interfaces

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// TransactionRepository defines the contract for transaction persistence operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error)
	// CreateTransfer atomically persists a debit/credit pair representing a
	// single transfer. The credit's ParentID is set to the debit's ID by the
	// implementation. Both records share the same Date.
	CreateTransfer(ctx context.Context, debit, credit domain.Transaction) (domain.Transaction, domain.Transaction, error)
	GetByID(ctx context.Context, id uint) (domain.Transaction, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Transaction, error)
	GetByUserIDPaginated(ctx context.Context, userID uint, limit, offset int) ([]domain.Transaction, error)
	GetByUserIDPaginatedWithFilters(ctx context.Context, userID uint, limit, offset int, startDate, endDate *time.Time) ([]domain.Transaction, error)
	GetByUserIDPaginatedWithAllFilters(ctx context.Context, userID uint, limit, offset int, txType *domain.TransactionType, accountID, categoryID *uint, startDate, endDate *time.Time) ([]domain.Transaction, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	CountByUserIDWithFilters(ctx context.Context, userID uint, startDate, endDate *time.Time) (int64, error)
	CountByUserIDWithAllFilters(ctx context.Context, userID uint, txType *domain.TransactionType, accountID, categoryID *uint, startDate, endDate *time.Time) (int64, error)
	GetByAccountID(ctx context.Context, accountID uint) ([]domain.Transaction, error)
	GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Transaction, error)
	GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Transaction, error)
	GetByType(ctx context.Context, userID uint, transactionType domain.TransactionType) ([]domain.Transaction, error)
	Update(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error)
	Delete(ctx context.Context, id uint) error
	ExistsByID(ctx context.Context, id uint) (bool, error)
}
