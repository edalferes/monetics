package user

import (
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/errors"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type RegisterUseCase struct {
	User            interfaces.User
	Role            interfaces.Role
	PasswordService interfaces.PasswordService
	logger          logger.Logger
}

func NewRegisterUseCase(user interfaces.User, role interfaces.Role, passwordService interfaces.PasswordService, log logger.Logger) *RegisterUseCase {
	return &RegisterUseCase{
		User:            user,
		Role:            role,
		PasswordService: passwordService,
		logger:          log.With().Str("usecase", "auth.register").Logger(),
	}
}

func (u *RegisterUseCase) Execute(username, password string) error {
	u.logger.Debug().Str("username", username).Msg("registering new user")

	if err := domain.DefaultPasswordPolicy().Validate(password); err != nil {
		u.logger.Warn().Err(err).Str("username", username).Msg("registration failed: weak password")
		return err
	}

	if user, _ := u.User.FindByUsername(username); user != nil {
		u.logger.Warn().Str("username", username).Msg("registration failed: user already exists")
		return errors.ErrUserAlreadyExists
	}
	hash, err := u.PasswordService.Hash(password)
	if err != nil {
		u.logger.Error().Err(err).Str("username", username).Msg("failed to hash password")
		return err
	}
	role, err := u.Role.FindByName("user")
	if err != nil {
		u.logger.Error().Err(err).Msg("failed to find default user role")
		return err
	}
	user := &domain.User{
		Username: username,
		Password: hash,
		Roles:    []domain.Role{*role},
	}
	err = u.User.Create(user)
	if err != nil {
		u.logger.Error().Err(err).Str("username", username).Msg("failed to create user")
		return err
	}
	u.logger.Info().Str("username", username).Uint("user_id", user.ID).Msg("user registered successfully")
	return nil
}
