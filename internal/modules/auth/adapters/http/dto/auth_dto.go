package dto

// LoginDTO is the request payload for user login.
type LoginDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the response payload for a successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// MessageResponse is a generic success message envelope.
type MessageResponse struct {
	Message string `json:"message"`
}

// AssignRoleRequest is the request payload to assign a role to a user.
type AssignRoleRequest struct {
	RoleName string `json:"role_name" validate:"required"`
}

// CreateRoleRequest is the request payload to create a role.
type CreateRoleRequest struct {
	Name          string `json:"name" validate:"required"`
	PermissionIDs []uint `json:"permission_ids,omitempty"`
}

// CreatePermissionRequest is the request payload to create a permission.
type CreatePermissionRequest struct {
	Name string `json:"name" validate:"required"`
}
