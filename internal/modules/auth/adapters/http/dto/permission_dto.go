package dto

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// PermissionDTO is the request payload for creating a permission.
type PermissionDTO struct {
	Name string `json:"name" validate:"required"`
}

// PermissionResponse represents a permission in API responses.
type PermissionResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ToPermissionResponse converts a domain.Permission to its API representation.
func ToPermissionResponse(p domain.Permission) PermissionResponse {
	return PermissionResponse{ID: p.ID, Name: p.Name}
}

// ToPermissionResponseList converts a slice of domain.Permission.
func ToPermissionResponseList(perms []domain.Permission) []PermissionResponse {
	out := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		out[i] = ToPermissionResponse(p)
	}
	return out
}
