package model

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// PermissionModel is the GORM persistence model for domain.Permission.
//
// Table: permissions.
type PermissionModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

func (PermissionModel) TableName() string { return "permissions" }

func (m PermissionModel) ToDomain() domain.Permission {
	return domain.Permission{
		ID:   m.ID,
		Name: m.Name,
	}
}

func PermissionFromDomain(p *domain.Permission) *PermissionModel {
	if p == nil {
		return nil
	}
	return &PermissionModel{
		ID:   p.ID,
		Name: p.Name,
	}
}

func PermissionsFromDomain(ps []domain.Permission) []PermissionModel {
	out := make([]PermissionModel, len(ps))
	for i, p := range ps {
		out[i] = PermissionModel{ID: p.ID, Name: p.Name}
	}
	return out
}

func PermissionModelsToDomain(ms []PermissionModel) []domain.Permission {
	out := make([]domain.Permission, len(ms))
	for i, m := range ms {
		out[i] = m.ToDomain()
	}
	return out
}
