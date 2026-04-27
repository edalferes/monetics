package dto

import "github.com/edalferes/monetics/internal/modules/auth/domain"

// RegisterDTO is the request payload for user registration / admin user create.
type RegisterDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserDTO is the request payload for updating a user.
type UpdateUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordDTO is the request payload for changing the current user's password.
type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// UserResponse represents a user in API responses (no password leaks).
type UserResponse struct {
	ID       uint           `json:"id"`
	Username string         `json:"username"`
	Roles    []RoleResponse `json:"roles,omitempty"`
}

// ToUserResponse converts a domain.User to its API representation.
func ToUserResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Roles:    ToRoleResponseList(u.Roles),
	}
}

// ToUserResponseList converts a slice of domain.User.
func ToUserResponseList(users []domain.User) []UserResponse {
	out := make([]UserResponse, len(users))
	for i, u := range users {
		out[i] = ToUserResponse(u)
	}
	return out
}
