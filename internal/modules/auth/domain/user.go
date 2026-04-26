// Package domain contains the core business entities for the auth module.
// All structs here are framework-agnostic: no GORM, JSON, or HTTP tags.
package domain

// User represents a system user with authentication and authorization data.
//
// Password stores a bcrypt hash, never plain text.
// Roles is the user's authorization set, populated by the repository.
type User struct {
	ID       uint
	Username string
	Password string
	Roles    []Role
}

// HasRole reports whether the user has a role by name.
func (u *User) HasRole(name string) bool {
	for _, r := range u.Roles {
		if r.Name == name {
			return true
		}
	}
	return false
}

// HasPermission reports whether the user has a permission via any role.
func (u *User) HasPermission(name string) bool {
	for _, r := range u.Roles {
		for _, p := range r.Permissions {
			if p.Name == name {
				return true
			}
		}
	}
	return false
}
