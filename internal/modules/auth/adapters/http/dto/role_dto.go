package dto

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// RoleDTO is the request payload for creating a role.
type RoleDTO struct {
	Name string `json:"name" validate:"required"`
}

// RoleResponse represents a role in API responses.
type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
}

// ToRoleResponse converts a domain.Role to its API representation.
func ToRoleResponse(r domain.Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Permissions: ToPermissionResponseList(r.Permissions),
	}
}

// ToRoleResponseList converts a slice of domain.Role.
func ToRoleResponseList(roles []domain.Role) []RoleResponse {
	out := make([]RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = ToRoleResponse(r)
	}
	return out
}
