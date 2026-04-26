package repository

import (
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
)

// RoleRepository is a GORM-backed implementation of interfaces.Role.
type RoleRepository struct {
	DB *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{DB: db}
}

var _ interfaces.Role = (*RoleRepository)(nil)

func (r *RoleRepository) FindByID(id uint) (*domain.Role, error) {
	var m model.RoleModel
	if err := r.DB.Preload("Permissions").First(&m, id).Error; err != nil {
		return nil, err
	}
	d := m.ToDomain()
	return &d, nil
}

func (r *RoleRepository) FindByName(name string) (*domain.Role, error) {
	var m model.RoleModel
	if err := r.DB.Preload("Permissions").Where("name = ?", name).First(&m).Error; err != nil {
		return nil, err
	}
	d := m.ToDomain()
	return &d, nil
}

func (r *RoleRepository) Create(role *domain.Role) error {
	m := model.RoleFromDomain(role)
	if err := r.DB.Create(m).Error; err != nil {
		return err
	}
	role.ID = m.ID
	return nil
}

func (r *RoleRepository) ListAll() ([]domain.Role, error) {
	var ms []model.RoleModel
	if err := r.DB.Preload("Permissions").Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.RoleModelsToDomain(ms), nil
}

func (r *RoleRepository) DeleteByName(name string) error {
	return r.DB.Where("name = ?", name).Delete(&model.RoleModel{}).Error
}

func (r *RoleRepository) Update(role *domain.Role) error {
	m := model.RoleFromDomain(role)
	return r.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(m).Error
}
