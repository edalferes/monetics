package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// TransactionRepository is a GORM-backed implementation of interfaces.TransactionRepository.
type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) interfaces.TransactionRepository {
	return &TransactionRepository{db: db}
}

// preload returns a query scope with Account/Category preloaded.
func (r *TransactionRepository) preload(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Preload("Account").
		Preload("Category")
}

func (r *TransactionRepository) Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	m := model.TransactionFromDomain(transaction)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Transaction{}, err
	}
	if err := r.preload(ctx).First(&m, m.ID).Error; err != nil {
		return domain.Transaction{}, err
	}
	return m.ToDomain(), nil
}

// CreateTransfer persists a debit/credit pair atomically. The credit row's
// ParentID is automatically set to the debit's ID after the first insert.
func (r *TransactionRepository) CreateTransfer(ctx context.Context, debit, credit domain.Transaction) (domain.Transaction, domain.Transaction, error) {
	debitModel := model.TransactionFromDomain(debit)
	creditModel := model.TransactionFromDomain(credit)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&debitModel).Error; err != nil {
			return err
		}
		creditModel.ParentID = &debitModel.ID
		if err := tx.Create(&creditModel).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return domain.Transaction{}, domain.Transaction{}, err
	}

	if err := r.preload(ctx).First(&debitModel, debitModel.ID).Error; err != nil {
		return domain.Transaction{}, domain.Transaction{}, err
	}
	if err := r.preload(ctx).First(&creditModel, creditModel.ID).Error; err != nil {
		return domain.Transaction{}, domain.Transaction{}, err
	}
	return debitModel.ToDomain(), creditModel.ToDomain(), nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uint) (domain.Transaction, error) {
	var m model.TransactionModel
	if err := r.preload(ctx).First(&m, id).Error; err != nil {
		return domain.Transaction{}, err
	}
	return m.ToDomain(), nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.preload(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) GetByUserIDPaginated(ctx context.Context, userID uint, limit, offset int) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.preload(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Limit(limit).
		Offset(offset).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID uint) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) GetByCategoryID(ctx context.Context, categoryID uint) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Order("date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) GetByType(ctx context.Context, userID uint, transactionType domain.TransactionType) ([]domain.Transaction, error) {
	var ms []model.TransactionModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, string(transactionType)).
		Order("date DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) Update(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	m := model.TransactionFromDomain(transaction)
	res := r.db.WithContext(ctx).Model(&m).
		Select(
			"account_id", "category_id", "type", "amount", "description", "date",
			"status", "tags", "attachments", "is_recurring", "recurrence_rule",
			"recurrence_end", "destination_account_id", "transfer_fee",
		).
		Updates(m)
	if res.Error != nil {
		return domain.Transaction{}, res.Error
	}
	if err := r.preload(ctx).First(&m, m.ID).Error; err != nil {
		return domain.Transaction{}, err
	}
	return m.ToDomain(), nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.TransactionModel{}, id).Error
}

func (r *TransactionRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *TransactionRepository) GetByUserIDPaginatedWithFilters(ctx context.Context, userID uint, limit, offset int, startDate, endDate *time.Time) ([]domain.Transaction, error) {
	query := r.preload(ctx).Where("user_id = ?", userID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	var ms []model.TransactionModel
	if err := query.Order("date DESC").Limit(limit).Offset(offset).Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

func (r *TransactionRepository) CountByUserIDWithFilters(ctx context.Context, userID uint, startDate, endDate *time.Time) (int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("user_id = ?", userID)
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetByUserIDPaginatedWithAllFilters retrieves paginated transactions with all filters.
func (r *TransactionRepository) GetByUserIDPaginatedWithAllFilters(
	ctx context.Context,
	userID uint,
	limit, offset int,
	txType *domain.TransactionType,
	accountID, categoryID *uint,
	startDate, endDate *time.Time,
) ([]domain.Transaction, error) {
	query := r.preload(ctx).Where("user_id = ?", userID)
	if txType != nil {
		query = query.Where("type = ?", *txType)
	}
	if accountID != nil {
		query = query.Where("account_id = ?", *accountID)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	var ms []model.TransactionModel
	if err := query.Order("date DESC").Limit(limit).Offset(offset).Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.TransactionModelsToDomain(ms), nil
}

// CountByUserIDWithAllFilters counts transactions with all filters.
func (r *TransactionRepository) CountByUserIDWithAllFilters(
	ctx context.Context,
	userID uint,
	txType *domain.TransactionType,
	accountID, categoryID *uint,
	startDate, endDate *time.Time,
) (int64, error) {
	query := r.db.WithContext(ctx).
		Model(&model.TransactionModel{}).
		Where("user_id = ?", userID)
	if txType != nil {
		query = query.Where("type = ?", *txType)
	}
	if accountID != nil {
		query = query.Where("account_id = ?", *accountID)
	}
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if startDate != nil {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", endDate)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
