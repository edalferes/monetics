package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new GORM-based account repository
func NewAccountRepository(db *gorm.DB) interfaces.AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	if err := r.db.WithContext(ctx).Create(&account).Error; err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uint) (domain.Account, error) {
	var account domain.Account
	if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Account, error) {
	var accounts []domain.Account
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *AccountRepository) Update(ctx context.Context, account domain.Account) (domain.Account, error) {
	if err := r.db.WithContext(ctx).Save(&account).Error; err != nil {
		return domain.Account{}, err
	}
	return account, nil
}

func (r *AccountRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&domain.Account{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *AccountRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Account{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
