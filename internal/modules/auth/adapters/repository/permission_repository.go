package repository

import (
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
)

// PermissionRepository is a GORM-backed implementation of interfaces.PermissionRepository.
type PermissionRepository struct {
	DB *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{DB: db}
}

var _ interfaces.PermissionRepository = (*PermissionRepository)(nil)

func (r *PermissionRepository) FindByID(id uint) (*domain.Permission, error) {
	var m model.PermissionModel
	if err := r.DB.First(&m, id).Error; err != nil {
		return nil, err
	}
	d := m.ToDomain()
	return &d, nil
}

func (r *PermissionRepository) FindByName(name string) (*domain.Permission, error) {
	var m model.PermissionModel
	if err := r.DB.Where("name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	d := m.ToDomain()
	return &d, nil
}

func (r *PermissionRepository) Create(permission *domain.Permission) error {
	m := model.PermissionFromDomain(permission)
	if err := r.DB.Create(m).Error; err != nil {
		return err
	}
	permission.ID = m.ID
	return nil
}

func (r *PermissionRepository) ListAll() ([]domain.Permission, error) {
	var ms []model.PermissionModel
	if err := r.DB.Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.PermissionModelsToDomain(ms), nil
}

func (r *PermissionRepository) DeleteByName(name string) error {
	return r.DB.Where("name = ?", name).Delete(&model.PermissionModel{}).Error
}

func (r *PermissionRepository) Update(permission *domain.Permission) error {
	m := model.PermissionFromDomain(permission)
	return r.DB.Save(m).Error
}
