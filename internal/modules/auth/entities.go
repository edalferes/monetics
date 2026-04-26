package auth

import (
	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository/model"
)

// Entities returns all GORM persistence models for the auth module.
func Entities() []interface{} {
	return []interface{}{
		&model.UserModel{},
		&model.RoleModel{},
		&model.PermissionModel{},
		&model.AuditLogModel{},
	}
}
