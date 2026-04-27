package repository

import (
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository/model"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
)

// UserRepository is a GORM-backed implementation of interfaces.User.
type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

var _ interfaces.User = (*UserRepository)(nil)

func (r *UserRepository) FindByID(id uint) (*domain.User, error) {
	var m model.UserModel
	if err := r.DB.Preload("Roles.Permissions").Preload("Roles").First(&m, id).Error; err != nil {
		return nil, err
	}
	u := m.ToDomain()
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	var m model.UserModel
	if err := r.DB.Preload("Roles.Permissions").Preload("Roles").Where("username = ?", username).First(&m).Error; err != nil {
		return nil, err
	}
	u := m.ToDomain()
	return &u, nil
}

func (r *UserRepository) ListAll() ([]domain.User, error) {
	var ms []model.UserModel
	if err := r.DB.Preload("Roles.Permissions").Preload("Roles").Find(&ms).Error; err != nil {
		return nil, err
	}
	return model.UserModelsToDomain(ms), nil
}

func (r *UserRepository) Create(user *domain.User) error {
	m := model.UserFromDomain(user)
	if err := r.DB.Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (r *UserRepository) Update(user *domain.User) error {
	m := model.UserFromDomain(user)
	return r.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(m).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.DB.Delete(&model.UserModel{}, id).Error
}
