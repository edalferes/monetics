package domain

// Role groups permissions together for role-based access control (RBAC).
type Role struct {
	ID          uint
	Name        string
	Permissions []Permission
}

// HasPermission reports whether the role has a permission by name.
func (r *Role) HasPermission(name string) bool {
	for _, p := range r.Permissions {
		if p.Name == name {
			return true
		}
	}
	return false
}
