package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// BudgetRepository is a GORM-backed implementation of interfaces.BudgetRepository.
type BudgetRepository struct {
	db *gorm.DB
}

func NewBudgetRepository(db *gorm.DB) interfaces.BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, budget domain.Budget) (domain.Budget, error) {
	m := model.BudgetFromDomain(budget)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Budget{}, err
	}
	return m.ToDomain(), nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, id uint) (domain.Budget, error) {
	var m model.BudgetModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return domain.Budget{}, err
	}
	return m.ToDomain(), nil
}

func (r *BudgetRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Budget, error) {
	var ms []model.BudgetModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.BudgetModelsToDomain(ms), nil
}

func (r *BudgetRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Budget, error) {
	var ms []model.BudgetModel
	if err := r.db.WithContext(ctx).Where("category_id = ?", categoryID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.BudgetModelsToDomain(ms), nil
}

func (r *BudgetRepository) GetActive(ctx context.Context, userID uint) ([]domain.Budget, error) {
	var ms []model.BudgetModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.BudgetModelsToDomain(ms), nil
}

func (r *BudgetRepository) Update(ctx context.Context, budget domain.Budget) (domain.Budget, error) {
	m := model.BudgetFromDomain(budget)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return domain.Budget{}, err
	}
	return m.ToDomain(), nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.BudgetModel{}, id).Error
}

func (r *BudgetRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.BudgetModel{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *BudgetRepository) UpdateSpent(ctx context.Context, budgetID uint, spent float64) error {
	return r.db.WithContext(ctx).
		Model(&model.BudgetModel{}).
		Where("id = ?", budgetID).
		Update("spent", spent).Error
}

func (r *BudgetRepository) GetOverlapping(ctx context.Context, userID uint, categoryID uint, startDate, endDate time.Time, excludeBudgetID *uint) ([]domain.Budget, error) {
	var ms []model.BudgetModel
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND category_id = ? AND is_active = ? AND start_date < ? AND end_date > ?",
			userID, categoryID, true, endDate, startDate)
	if excludeBudgetID != nil {
		query = query.Where("id != ?", *excludeBudgetID)
	}
	if err := query.Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.BudgetModelsToDomain(ms), nil
}

// UpdateSpentAtomic recomputes budget.spent from the underlying transactions in a single SQL update.
func (r *BudgetRepository) UpdateSpentAtomic(ctx context.Context, budgetID uint, categoryID uint, startDate, endDate time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.BudgetModel{}).
		Where("id = ?", budgetID).
		Update("spent", r.db.Raw(
			"COALESCE((SELECT SUM(amount) FROM budget_transactions WHERE category_id = ? AND type = ? AND date BETWEEN ? AND ? AND status = ?), 0)",
			categoryID, string(domain.TransactionTypeExpense), startDate, endDate, string(domain.TransactionStatusCompleted),
		)).Error
}
