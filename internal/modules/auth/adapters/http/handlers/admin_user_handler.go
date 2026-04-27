package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/errors"
)

type AdminUserHandler struct {
	ListUsersUC  interface{ Execute() ([]domain.User, error) }
	CreateUserUC interface {
		Execute(username, password string) error
	}
	GetUserByIDUC interface {
		Execute(id uint) (*domain.User, error)
	}
	UpdateUserUC interface {
		Execute(userID uint, username, password string) error
	}
	DeleteUserUC interface{ Execute(id uint) error }
	AssignRoleUC interface {
		Execute(userID uint, roleName string) error
	}
	RemoveRoleUC interface {
		Execute(userID uint, roleName string) error
	}
}

// ListUsers godoc
// @Summary List all users
// @Tags Auth - Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} domain.User
// @Router /v1/admin/users [get]
func (h *AdminUserHandler) ListUsers(c echo.Context) error {
	users, err := h.ListUsersUC.Execute()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, dto.ToUserResponseList(users))
}

// CreateUser godoc
// @Summary Create a new user
// @Tags Auth - Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body dto.RegisterDTO true "User data"
// @Success 201 {object} map[string]string "created"
// @Failure 400 {object} map[string]string "invalid data or user exists"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users [post]
func (h *AdminUserHandler) CreateUser(c echo.Context) error {
	var input dto.RegisterDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidData.Error()})
	}
	if err := validator.New().Struct(input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errors.ErrMissingCredentials.Error()})
	}
	if err := h.CreateUserUC.Execute(input.Username, input.Password); err != nil {
		if err == errors.ErrUserAlreadyExists {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
	return c.JSON(http.StatusCreated, map[string]string{"message": "user created"})
}

// GetUser godoc
// @Summary Get user by ID
// @Tags Auth - Admin
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string "invalid user id"
// @Failure 404 {object} map[string]string "user not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users/{id} [get]
func (h *AdminUserHandler) GetUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	userID := uint(id)

	user, err := h.GetUserByIDUC.Execute(userID)
	if err != nil {
		if err == errors.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, dto.ToUserResponse(*user))
}

// UpdateUser godoc
// @Summary Update user by ID
// @Tags Auth - Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.RegisterDTO true "User data"
// @Success 200 {object} map[string]string "updated"
// @Failure 400 {object} map[string]string "invalid data"
// @Failure 404 {object} map[string]string "user not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users/{id} [put]
func (h *AdminUserHandler) UpdateUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	userID := uint(id)

	var input dto.UpdateUserDTO
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidData.Error()})
	}

	if err := h.UpdateUserUC.Execute(userID, input.Username, input.Password); err != nil {
		if err == errors.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user updated"})
}

// DeleteUser godoc
// @Summary Delete user by ID
// @Tags Auth - Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "invalid user id"
// @Failure 404 {object} map[string]string "user not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users/{id} [delete]
func (h *AdminUserHandler) DeleteUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	userID := uint(id)

	if err := h.DeleteUserUC.Execute(userID); err != nil {
		if err == errors.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.NoContent(http.StatusNoContent)
}

// AssignRoleToUser godoc
// @Summary Assign role to user
// @Tags Auth - Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param role body map[string]string true "Role name"
// @Success 200 {object} map[string]string "role assigned"
// @Failure 400 {object} map[string]string "invalid data"
// @Failure 404 {object} map[string]string "user or role not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users/{id}/roles [post]
func (h *AdminUserHandler) AssignRoleToUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	userID := uint(id)

	var req dto.AssignRoleRequest
	if err := c.Bind(&req); err != nil || req.RoleName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "role_name is required"})
	}

	if err := h.AssignRoleUC.Execute(userID, req.RoleName); err != nil {
		if err == errors.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "role assigned"})
}

// RemoveRoleFromUser godoc
// @Summary Remove role from user
// @Tags Auth - Admin
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param roleName path string true "Role name"
// @Success 200 {object} map[string]string "role removed"
// @Failure 400 {object} map[string]string "invalid user id"
// @Failure 404 {object} map[string]string "user or role not found"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/admin/users/{id}/roles/{roleName} [delete]
func (h *AdminUserHandler) RemoveRoleFromUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}
	userID := uint(id)

	roleName := c.Param("roleName")
	if roleName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "role name is required"})
	}

	if err := h.RemoveRoleUC.Execute(userID, roleName); err != nil {
		if err == errors.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "role removed"})
}
