package auth

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/crypto"
	"github.com/edalferes/monetics/internal/modules/auth/adapters/repository"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
)

func Seed(db *gorm.DB, adminUsername, adminPassword string) error {
	if strings.TrimSpace(adminUsername) == "" {
		return fmt.Errorf("admin username is required for auth seed")
	}
	if strings.TrimSpace(adminPassword) == "" {
		return fmt.Errorf("admin password is required for auth seed")
	}

	roleRepo := repository.NewRoleRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	userRepo := repository.NewUserRepository(db)
	passwordService := crypto.NewBcryptPasswordService()

	defaultRoles := []string{"admin", "user"}
	defaultPerms := []string{"read", "write", "delete"}

	// Seed permissions
	for _, permName := range defaultPerms {
		_, err := permRepo.FindByName(permName)
		if err != nil {
			if err := permRepo.Create(&domain.Permission{Name: permName}); err != nil {
				return err
			}
		}
	}

	// Seed roles and assign permissions
	allPerms, err := permRepo.ListAll()
	if err != nil {
		return err
	}

	for _, roleName := range defaultRoles {
		var permsToAssign []domain.Permission
		if roleName == "admin" {
			permsToAssign = allPerms // admin allows all permissions
		} else {
			// user only read
			for _, p := range allPerms {
				if p.Name == "read" {
					permsToAssign = append(permsToAssign, p)
				}
			}
		}

		// Check if role already exists
		_, err := roleRepo.FindByName(roleName)
		if err != nil {
			// Role doesn't exist, create it
			role := &domain.Role{Name: roleName, Permissions: permsToAssign}
			if err := roleRepo.Create(role); err != nil {
				return err
			}
		}
		// Note: For simplicity, we're not updating existing roles permissions
		// In a real app, you might want to implement an Update method for this
	}

	// Seed admin user
	_, err = userRepo.FindByUsername(adminUsername)
	if err != nil {
		adminRole, err := roleRepo.FindByName("admin")
		if err != nil {
			return err
		}
		userRole, err := roleRepo.FindByName("user")
		if err != nil {
			return err
		}
		hash, err := passwordService.Hash(adminPassword)
		if err != nil {
			return err
		}
		adminUser := &domain.User{
			Username: adminUsername,
			Password: hash,
			Roles:    []domain.Role{*adminRole, *userRole},
		}
		if err := userRepo.Create(adminUser); err != nil {
			return err
		}
	}
	return nil
}
