package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// CategoryRepository is a GORM-backed implementation of interfaces.CategoryRepository.
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) interfaces.CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	m := model.CategoryFromDomain(category)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return domain.Category{}, err
	}
	return m.ToDomain(), nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (domain.Category, error) {
	var m model.CategoryModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return domain.Category{}, err
	}
	return m.ToDomain(), nil
}

func (r *CategoryRepository) GetByUserID(ctx context.Context, userID uint) ([]domain.Category, error) {
	var ms []model.CategoryModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.CategoryModelsToDomain(ms), nil
}

func (r *CategoryRepository) GetByType(ctx context.Context, userID uint, categoryType domain.CategoryType) ([]domain.Category, error) {
	var ms []model.CategoryModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, string(categoryType)).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.CategoryModelsToDomain(ms), nil
}

func (r *CategoryRepository) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	m := model.CategoryFromDomain(category)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return domain.Category{}, err
	}
	return m.ToDomain(), nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.CategoryModel{}, id).Error
}

func (r *CategoryRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CategoryModel{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}
