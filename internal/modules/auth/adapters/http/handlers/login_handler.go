// Package login provides HTTP handlers for user authentication endpoints.
//
// This package implements the HTTP adapter layer for authentication operations,
// specifically handling login requests and JWT token generation. It follows
// Clean Architecture principles by delegating business logic to use cases.
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/auth/errors"
	userUC "github.com/edalferes/monetics/internal/modules/auth/usecase/user"
	"github.com/edalferes/monetics/pkg/logger"
	"github.com/edalferes/monetics/pkg/utils"
)

// Handler handles HTTP requests for user authentication.
//
// This handler is responsible for:
//   - Binding HTTP request data to DTOs
//   - Delegating business logic to appropriate use cases
//   - Converting use case results to HTTP responses
//   - Logging HTTP-layer events and errors
//   - Returning appropriate HTTP status codes
//
// The handler follows Clean Architecture principles by:
//   - Depending on use case interfaces, not implementations
//   - Converting between HTTP and domain models
//   - Handling HTTP-specific concerns (status codes, headers)
//   - Delegating all business logic to use cases
//
// Example usage:
//
//	handler := &LoginHandler{
//		LoginUseCase: loginUseCase,
//		Logger:       logger,
//	}
//
//	e.POST("/auth/login", handler.Login)
type LoginHandler struct {
	LoginUseCase *userUC.LoginWithAuditUseCase // Use case for authentication logic
	Logger       logger.Logger                 // Logger for HTTP-layer events
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         Auth - Login
// @Accept json
// @Produce json
// @Param credentials body dto.LoginDTO true "User credentials"
// @Success 200 {object} map[string]string "token"
// @Failure 400 {object} map[string]string "invalid data or missing credentials"
// @Failure 401 {object} map[string]string "invalid credentials"
// @Failure 500 {object} map[string]string "internal error"
// @Router /v1/auth/login [post]
func (h *LoginHandler) Login(c echo.Context) error {
	var input dto.LoginDTO
	if err := c.Bind(&input); err != nil {
		h.Logger.Error().Err(err).Msg("failed to bind login request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errors.ErrInvalidData.Error()})
	}
	if input.Username == "" || input.Password == "" {
		h.Logger.Warn().Str("username", input.Username).Msg("login attempt with missing credentials")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": errors.ErrMissingCredentials.Error()})
	}
	ip := utils.ToIPv4(c.RealIP())
	h.Logger.Info().Str("username", input.Username).Str("ip", ip).Msg("login attempt")

	token, err := h.LoginUseCase.Execute(input.Username, input.Password, ip)
	if err != nil {
		if err == errors.ErrInvalidCredentials {
			h.Logger.Warn().Str("username", input.Username).Str("ip", ip).Msg("invalid credentials")
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		h.Logger.Error().Err(err).Str("username", input.Username).Msg("internal error during login")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}

	h.Logger.Info().Str("username", input.Username).Str("ip", ip).Msg("successful login")
	return c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
}
