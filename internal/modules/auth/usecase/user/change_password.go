package user

import (
	"github.com/edalferes/monetics/internal/modules/auth/domain"
	"github.com/edalferes/monetics/internal/modules/auth/errors"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type ChangePasswordUseCase struct {
	UserRepo        interfaces.User
	PasswordService interfaces.PasswordService
	logger          logger.Logger
}

func NewChangePasswordUseCase(userRepo interfaces.User, passwordService interfaces.PasswordService, log logger.Logger) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		UserRepo:        userRepo,
		PasswordService: passwordService,
		logger:          log.With().Str("usecase", "auth.change_password").Logger(),
	}
}

// Execute changes the user's password after verifying the current password
func (u *ChangePasswordUseCase) Execute(userID uint, currentPassword, newPassword string) error {
	u.logger.Debug().Uint("user_id", userID).Msg("changing password")

	if err := domain.DefaultPasswordPolicy().Validate(newPassword); err != nil {
		u.logger.Warn().Err(err).Uint("user_id", userID).Msg("password change failed: weak password")
		return err
	}

	// Find the user
	user, err := u.UserRepo.FindByID(userID)
	if err != nil {
		u.logger.Error().Err(err).Uint("user_id", userID).Msg("user not found")
		return errors.ErrUserNotFound
	}

	// Verify current password
	if err := u.PasswordService.Compare(user.Password, currentPassword); err != nil {
		u.logger.Warn().Uint("user_id", userID).Msg("password change failed: invalid current password")
		return errors.ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := u.PasswordService.Hash(newPassword)
	if err != nil {
		u.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to hash new password")
		return errors.ErrInvalidData
	}

	// Update user password
	user.Password = hashedPassword
	err = u.UserRepo.Update(user)
	if err != nil {
		u.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to update password")
		return err
	}

	u.logger.Info().Uint("user_id", userID).Msg("password changed successfully")
	return nil
}
