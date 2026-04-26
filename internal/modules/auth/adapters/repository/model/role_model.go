package model

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// RoleModel is the GORM persistence model for domain.Role.
//
// Table: roles. Permissions is a many-to-many through role_permissions.
type RoleModel struct {
	ID          uint              `gorm:"primaryKey"`
	Name        string            `gorm:"unique;not null"`
	Permissions []PermissionModel `gorm:"many2many:role_permissions;joinForeignKey:role_model_id;joinReferences:permission_model_id;constraint:OnDelete:CASCADE"`
}

func (RoleModel) TableName() string { return "roles" }

func (m RoleModel) ToDomain() domain.Role {
	r := domain.Role{
		ID:   m.ID,
		Name: m.Name,
	}
	if len(m.Permissions) > 0 {
		r.Permissions = PermissionModelsToDomain(m.Permissions)
	}
	return r
}

func RoleFromDomain(r *domain.Role) *RoleModel {
	if r == nil {
		return nil
	}
	m := &RoleModel{
		ID:   r.ID,
		Name: r.Name,
	}
	if len(r.Permissions) > 0 {
		m.Permissions = PermissionsFromDomain(r.Permissions)
	}
	return m
}

func RolesFromDomain(rs []domain.Role) []RoleModel {
	out := make([]RoleModel, len(rs))
	for i, r := range rs {
		rr := r
		out[i] = *RoleFromDomain(&rr)
	}
	return out
}

func RoleModelsToDomain(ms []RoleModel) []domain.Role {
	out := make([]domain.Role, len(ms))
	for i, m := range ms {
		out[i] = m.ToDomain()
	}
	return out
}
