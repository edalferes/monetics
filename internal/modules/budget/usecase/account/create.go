package account

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// CreateUseCase handles account creation
type CreateUseCase struct {
	accountRepo interfaces.AccountRepository
	logger      logger.Logger
}

// NewCreateUseCase creates a new use case instance
func NewCreateUseCase(accountRepo interfaces.AccountRepository, log logger.Logger) *CreateUseCase {
	return &CreateUseCase{
		accountRepo: accountRepo,
		logger:      log.With().Str("usecase", "account.create").Logger(),
	}
}

// CreateInput represents the input for creating an account
type CreateInput struct {
	UserID      uint               `json:"user_id" validate:"required"`
	Name        string             `json:"name" validate:"required"`
	Type        domain.AccountType `json:"type" validate:"required"`
	Currency    string             `json:"currency"`
	Description string             `json:"description,omitempty"`
}

// Execute creates a new account
func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (domain.Account, error) {
	uc.logger.Debug().
		Uint("user_id", input.UserID).
		Str("name", input.Name).
		Str("type", string(input.Type)).
		Msg("creating account")

	// Validate input
	if input.UserID == 0 {
		uc.logger.Error().Msg("invalid user_id: cannot be zero")
		return domain.Account{}, errors.ErrInvalidUserID
	}
	if input.Name == "" {
		uc.logger.Error().Msg("account name is required")
		return domain.Account{}, errors.ErrAccountNameRequired
	}
	if !isValidAccountType(input.Type) {
		uc.logger.Error().Str("type", string(input.Type)).Msg("invalid account type")
		return domain.Account{}, errors.ErrInvalidAccountType
	}
	if input.Currency == "" {
		input.Currency = "BRL" // Default currency
	}

	// Create account domain entity
	account := domain.Account{
		UserID:      input.UserID,
		Name:        input.Name,
		Type:        input.Type,
		Balance:     0,
		Currency:    input.Currency,
		Description: input.Description,
		IsActive:    true,
	}

	// Save to repository
	createdAccount, err := uc.accountRepo.Create(ctx, account)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", input.UserID).Msg("failed to create account")
		return domain.Account{}, err
	}

	uc.logger.Info().
		Uint("account_id", createdAccount.ID).
		Uint("user_id", createdAccount.UserID).
		Str("name", createdAccount.Name).
		Msg("account created successfully")

	return createdAccount, nil
}
