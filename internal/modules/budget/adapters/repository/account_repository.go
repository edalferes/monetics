package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// AccountRepository is a GORM-backed implementation of interfaces.AccountRepository.
// It converts between domain.Account and model.AccountModel at the boundary.
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new GORM-based account repository.
func NewAccountRepository(db *gorm.DB) interfaces.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	m := model.AccountFromDomain(account)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Account{}, err
	}
	return m.ToDomain(), nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uint) (domain.Account, error) {
	var m model.AccountModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return domain.Account{}, err
	}
	return m.ToDomain(), nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Account, error) {
	var ms []model.AccountModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.AccountModelsToDomain(ms), nil
}

func (r *AccountRepository) Update(ctx context.Context, account domain.Account) (domain.Account, error) {
	m := model.AccountFromDomain(account)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return domain.Account{}, err
	}
	return m.ToDomain(), nil
}

// Delete performs a soft delete by marking the account as inactive.
func (r *AccountRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.AccountModel{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *AccountRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.AccountModel{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}
