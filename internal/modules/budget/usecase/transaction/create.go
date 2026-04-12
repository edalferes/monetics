package transaction

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/helpers"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

type CreateUseCase struct {
	transactionRepo interfaces.TransactionRepository
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

func NewCreateUseCase(transactionRepo interfaces.TransactionRepository, accountRepo interfaces.AccountRepository, categoryRepo interfaces.CategoryRepository, budgetRepo interfaces.BudgetRepository, log logger.Logger) *CreateUseCase {
	return &CreateUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "transaction.create").Logger(),
	}
}

type CreateInput struct {
	UserID               uint
	AccountID            uint
	CategoryID           uint
	Type                 domain.TransactionType
	Amount               float64
	Description          string
	Date                 string
	DestinationAccountID *uint
}

func (uc *CreateUseCase) Execute(ctx context.Context, input CreateInput) (domain.Transaction, error) {
	uc.logger.Debug().
		Uint("user_id", input.UserID).
		Uint("account_id", input.AccountID).
		Uint("category_id", input.CategoryID).
		Str("type", string(input.Type)).
		Msg("creating transaction")

	if input.UserID == 0 {
		uc.logger.Error().Msg("invalid user_id: cannot be zero")
		return domain.Transaction{}, errors.ErrInvalidUserID
	}
	if input.Amount <= 0 {
		uc.logger.Error().Msg("invalid amount: must be positive")
		return domain.Transaction{}, errors.ErrInvalidAmount
	}
	if !isValidTransactionType(input.Type) {
		uc.logger.Error().Str("type", string(input.Type)).Msg("invalid transaction type")
		return domain.Transaction{}, errors.ErrInvalidTransactionType
	}

	account, err := uc.accountRepo.GetByID(ctx, input.AccountID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("account_id", input.AccountID).Msg("account not found")
		return domain.Transaction{}, errors.ErrAccountNotFound
	}
	if account.UserID != input.UserID {
		uc.logger.Warn().
			Uint("account_user_id", account.UserID).
			Uint("request_user_id", input.UserID).
			Msg("unauthorized access: account belongs to different user")
		return domain.Transaction{}, errors.ErrUnauthorizedAccess
	}

	category, err := uc.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("category_id", input.CategoryID).Msg("category not found")
		return domain.Transaction{}, errors.ErrCategoryNotFound
	}
	if category.UserID != input.UserID {
		uc.logger.Warn().
			Uint("category_user_id", category.UserID).
			Uint("request_user_id", input.UserID).
			Msg("unauthorized access: category belongs to different user")
		return domain.Transaction{}, errors.ErrUnauthorizedAccess
	}

	// Validate transfer-specific fields
	var destinationAccountID *uint
	if input.Type == domain.TransactionTypeTransfer {
		// Transfer requires a destination account
		if input.DestinationAccountID == nil || *input.DestinationAccountID == 0 {
			uc.logger.Error().Msg("transfer requires destination account")
			return domain.Transaction{}, errors.ErrTransferRequiresDestination
		}
		// Cannot transfer to same account
		if *input.DestinationAccountID == input.AccountID {
			uc.logger.Error().Msg("cannot transfer to same account")
			return domain.Transaction{}, errors.ErrTransferSameAccount
		}
		destinationAccountID = input.DestinationAccountID

		// Validate destination account exists and belongs to user
		uc.logger.Debug().Uint("destination_account_id", *input.DestinationAccountID).Msg("validating destination account for transfer")
		destAccount, err := uc.accountRepo.GetByID(ctx, *input.DestinationAccountID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("destination_account_id", *input.DestinationAccountID).Msg("destination account not found")
			return domain.Transaction{}, errors.ErrAccountNotFound
		}
		if destAccount.UserID != input.UserID {
			uc.logger.Warn().
				Uint("dest_account_user_id", destAccount.UserID).
				Uint("request_user_id", input.UserID).
				Msg("unauthorized access: destination account belongs to different user")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
	} else if input.DestinationAccountID != nil && *input.DestinationAccountID > 0 {
		// Non-transfer transactions should not have destination account; normalize to nil
		destinationAccountID = nil
	}

	// Parse date
	date, err := helpers.ParseFlexibleDate(input.Date)
	if err != nil {
		uc.logger.Error().Err(err).Str("date", input.Date).Msg("invalid date format")
		return domain.Transaction{}, errors.ErrInvalidDate
	}

	// Create the main transaction (debit from source account)
	tx := domain.Transaction{
		UserID:               input.UserID,
		AccountID:            input.AccountID,
		CategoryID:           input.CategoryID,
		Type:                 input.Type,
		Amount:               input.Amount,
		Description:          input.Description,
		Date:                 date,
		Month:                date.Format("2006-01"),
		Status:               domain.TransactionStatusCompleted,
		DestinationAccountID: destinationAccountID,
	}

	createdTx, err := uc.transactionRepo.Create(ctx, tx)
	if err != nil {
		uc.logger.Error().Err(err).
			Uint("user_id", input.UserID).
			Uint("account_id", input.AccountID).
			Msg("failed to create transaction")
		return domain.Transaction{}, err
	}

	uc.logger.Info().
		Uint("transaction_id", createdTx.ID).
		Uint("user_id", createdTx.UserID).
		Str("type", string(createdTx.Type)).
		Msg("transaction created successfully")

	// Update budget spent if this is an expense transaction
	if input.Type == domain.TransactionTypeExpense {
		uc.logger.Debug().Uint("user_id", input.UserID).Uint("category_id", input.CategoryID).Msg("updating budget spent")
		uc.updateBudgetSpent(ctx, input.UserID, input.CategoryID, date)
	}

	return createdTx, nil
}

// updateBudgetSpent atomically updates the spent amount for active budgets of the category
func (uc *CreateUseCase) updateBudgetSpent(ctx context.Context, userID, categoryID uint, transactionDate time.Time) {
	uc.logger.Debug().Uint("user_id", userID).Uint("category_id", categoryID).Msg("updating budget spent")

	// Get active budgets for this category
	budgets, err := uc.budgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to get budgets for spent update")
		return
	}

	for _, budget := range budgets {
		// Check if budget is for this category and transaction date is within period
		if budget.CategoryID == categoryID &&
			budget.IsActive &&
			!transactionDate.Before(budget.StartDate) &&
			!transactionDate.After(budget.EndDate) {

			// Atomically update spent using SQL subquery
			err = uc.budgetRepo.UpdateSpentAtomic(ctx, budget.ID, categoryID, budget.StartDate, budget.EndDate)
			if err != nil {
				uc.logger.Error().Err(err).Uint("budget_id", budget.ID).Msg("failed to update budget spent")
			} else {
				uc.logger.Info().Uint("budget_id", budget.ID).Msg("budget spent updated successfully")
			}
		}
	}
}
