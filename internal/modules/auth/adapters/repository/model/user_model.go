package model

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// UserModel is the GORM persistence model for domain.User.
//
// Table: users. Roles is a many-to-many relationship through user_roles.
// The join column names (user_model_id, role_model_id) follow GORM's default
// derivation from the struct name.
type UserModel struct {
	ID       uint        `gorm:"primaryKey"`
	Username string      `gorm:"unique;not null"`
	Password string      `gorm:"not null;column:password"`
	Roles    []RoleModel `gorm:"many2many:user_roles;joinForeignKey:user_model_id;joinReferences:role_model_id;constraint:OnDelete:CASCADE"`
}

func (UserModel) TableName() string { return "users" }

func (m UserModel) ToDomain() domain.User {
	u := domain.User{
		ID:       m.ID,
		Username: m.Username,
		Password: m.Password,
	}
	if len(m.Roles) > 0 {
		u.Roles = RoleModelsToDomain(m.Roles)
	}
	return u
}

func UserFromDomain(u *domain.User) *UserModel {
	if u == nil {
		return nil
	}
	m := &UserModel{
		ID:       u.ID,
		Username: u.Username,
		Password: u.Password,
	}
	if len(u.Roles) > 0 {
		m.Roles = RolesFromDomain(u.Roles)
	}
	return m
}

func UserModelsToDomain(ms []UserModel) []domain.User {
	out := make([]domain.User, len(ms))
	for i, m := range ms {
		out[i] = m.ToDomain()
	}
	return out
}
