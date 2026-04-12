package interfaces

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// BudgetRepository defines the contract for budget persistence operations
type BudgetRepository interface {
	Create(ctx context.Context, budget domain.Budget) (domain.Budget, error)
	GetByID(ctx context.Context, id uint) (domain.Budget, error)
	GetByUserID(ctx context.Context, userID uint) ([]domain.Budget, error)
	GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Budget, error)
	GetActive(ctx context.Context, userID uint) ([]domain.Budget, error)
	GetOverlapping(ctx context.Context, userID uint, categoryID uint, startDate, endDate time.Time, excludeBudgetID *uint) ([]domain.Budget, error)
	Update(ctx context.Context, budget domain.Budget) (domain.Budget, error)
	Delete(ctx context.Context, id uint) error
	ExistsByID(ctx context.Context, id uint) (bool, error)
	UpdateSpent(ctx context.Context, budgetID uint, spent float64) error
	UpdateSpentAtomic(ctx context.Context, budgetID uint, categoryID uint, startDate, endDate time.Time) error
}
