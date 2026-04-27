package transaction

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// UpdateInput represents the input for updating a transaction.
//
// Transfer semantics: when the resulting type is "transfer" the use case
// rebuilds the debit/credit pair from scratch (the previous row and its pair
// are deleted, then CreateTransfer is called) so that the RN-T3 invariant
// (one transfer = two linked rows) is always preserved. The same replace
// strategy is applied when the type changes from "transfer" to a single-row
// type.
type UpdateInput struct {
	ID                   uint
	UserID               uint
	AccountID            *uint
	CategoryID           *uint
	Type                 *domain.TransactionType
	Amount               *float64
	Description          *string
	Date                 *string
	DestinationAccountID *uint
}

// UpdateUseCase handles transaction updates
type UpdateUseCase struct {
	transactionRepo interfaces.TransactionRepository
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	budgetRepo      interfaces.BudgetRepository
	logger          logger.Logger
}

// NewUpdateUseCase creates a new use case instance
func NewUpdateUseCase(
	transactionRepo interfaces.TransactionRepository,
	accountRepo interfaces.AccountRepository,
	categoryRepo interfaces.CategoryRepository,
	budgetRepo interfaces.BudgetRepository,
	log logger.Logger,
) *UpdateUseCase {
	return &UpdateUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		budgetRepo:      budgetRepo,
		logger:          log.With().Str("usecase", "transaction.update").Logger(),
	}
}

// Execute updates an existing transaction.
func (uc *UpdateUseCase) Execute(ctx context.Context, input UpdateInput) (domain.Transaction, error) {
	uc.logger.Debug().Uint("transaction_id", input.ID).Uint("user_id", input.UserID).Msg("updating transaction")

	if input.ID == 0 {
		uc.logger.Error().Msg("invalid transaction_id: cannot be zero")
		return domain.Transaction{}, errors.ErrTransactionNotFound
	}

	tx, err := uc.transactionRepo.GetByID(ctx, input.ID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", input.ID).Msg("transaction not found")
		return domain.Transaction{}, err
	}

	if tx.UserID != input.UserID {
		uc.logger.Warn().
			Uint("transaction_user_id", tx.UserID).
			Uint("request_user_id", input.UserID).
			Msg("unauthorized access: transaction belongs to different user")
		return domain.Transaction{}, errors.ErrTransactionNotFound
	}

	// If the existing row is the credit leg of a transfer, redirect the update
	// to the debit leg so the resulting state is unambiguous.
	if tx.Type == domain.TransactionTypeTransfer && tx.ParentID != nil {
		debit, err := uc.transactionRepo.GetByID(ctx, *tx.ParentID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("parent_id", *tx.ParentID).Msg("failed to load transfer debit leg")
			return domain.Transaction{}, err
		}
		tx = debit
		input.ID = debit.ID
	}

	if input.AccountID != nil {
		account, err := uc.accountRepo.GetByID(ctx, *input.AccountID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("account_id", *input.AccountID).Msg("account not found for update")
			return domain.Transaction{}, errors.ErrAccountNotFound
		}
		if account.UserID != input.UserID {
			uc.logger.Warn().Uint("account_user_id", account.UserID).Uint("request_user_id", input.UserID).Msg("unauthorized account access")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
	}

	if input.CategoryID != nil {
		category, err := uc.categoryRepo.GetByID(ctx, *input.CategoryID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("category_id", *input.CategoryID).Msg("category not found for update")
			return domain.Transaction{}, errors.ErrCategoryNotFound
		}
		if category.UserID != input.UserID {
			uc.logger.Warn().Uint("category_user_id", category.UserID).Uint("request_user_id", input.UserID).Msg("unauthorized category access")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
	}

	if input.Type != nil && !isValidTransactionType(*input.Type) {
		uc.logger.Error().Str("type", string(*input.Type)).Msg("invalid transaction type")
		return domain.Transaction{}, errors.ErrInvalidTransactionType
	}

	if input.Amount != nil && *input.Amount <= 0 {
		uc.logger.Error().Msg("invalid amount: must be positive")
		return domain.Transaction{}, errors.ErrInvalidAmount
	}

	// Compute merged target state.
	target := tx
	if input.AccountID != nil {
		target.AccountID = *input.AccountID
	}
	if input.CategoryID != nil {
		target.CategoryID = *input.CategoryID
	}
	if input.Type != nil {
		target.Type = *input.Type
	}
	if input.Amount != nil {
		target.Amount = *input.Amount
	}
	if input.Description != nil {
		target.Description = *input.Description
	}
	if input.Date != nil {
		parsedDate, err := time.Parse(time.RFC3339, *input.Date)
		if err != nil {
			uc.logger.Error().Err(err).Str("date", *input.Date).Msg("invalid date format")
			return domain.Transaction{}, errors.ErrInvalidDate
		}
		target.Date = parsedDate
		target.Month = parsedDate.Format("2006-01")
		if target.Status != domain.TransactionStatusCancelled {
			target.Status = domain.ResolveStatus(parsedDate, time.Now())
		}
	}
	if input.DestinationAccountID != nil {
		target.DestinationAccountID = input.DestinationAccountID
	}
	if target.Type != domain.TransactionTypeTransfer {
		target.DestinationAccountID = nil
	}

	wasTransfer := tx.Type == domain.TransactionTypeTransfer
	isTransfer := target.Type == domain.TransactionTypeTransfer

	affectedCategories := map[uint]bool{tx.CategoryID: true, target.CategoryID: true}

	var resultTx domain.Transaction

	if wasTransfer || isTransfer {
		resultTx, err = uc.replaceWithPair(ctx, tx, target)
		if err != nil {
			return domain.Transaction{}, err
		}
	} else {
		resultTx, err = uc.transactionRepo.Update(ctx, target)
		if err != nil {
			uc.logger.Error().Err(err).Uint("transaction_id", input.ID).Msg("failed to update transaction")
			return domain.Transaction{}, err
		}
	}

	for categoryID := range affectedCategories {
		uc.recalculateBudgetSpent(ctx, input.UserID, categoryID, resultTx.Date)
	}

	uc.logger.Info().Uint("transaction_id", resultTx.ID).Msg("transaction updated successfully")
	return resultTx, nil
}

// replaceWithPair handles updates that involve a transfer (either before or
// after the change). It deletes the original row plus its paired row (if any)
// and recreates the transaction as a single row or as a debit/credit pair.
func (uc *UpdateUseCase) replaceWithPair(
	ctx context.Context,
	original, target domain.Transaction,
) (domain.Transaction, error) {
	isTransfer := target.Type == domain.TransactionTypeTransfer

	if isTransfer {
		if target.DestinationAccountID == nil || *target.DestinationAccountID == 0 {
			uc.logger.Error().Msg("transfer requires destination account")
			return domain.Transaction{}, errors.ErrTransferRequiresDestination
		}
		if *target.DestinationAccountID == target.AccountID {
			uc.logger.Error().Msg("cannot transfer to same account")
			return domain.Transaction{}, errors.ErrTransferSameAccount
		}
		destAccount, err := uc.accountRepo.GetByID(ctx, *target.DestinationAccountID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("destination_account_id", *target.DestinationAccountID).Msg("destination account not found")
			return domain.Transaction{}, errors.ErrAccountNotFound
		}
		if destAccount.UserID != target.UserID {
			uc.logger.Warn().Msg("unauthorized destination account access")
			return domain.Transaction{}, errors.ErrUnauthorizedAccess
		}
	}

	// Delete the paired row of the original transfer (if any).
	if original.Type == domain.TransactionTypeTransfer {
		pair, ok, err := uc.transactionRepo.GetTransferPair(ctx, original.ID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("transaction_id", original.ID).Msg("failed to fetch transfer pair")
			return domain.Transaction{}, err
		}
		if ok {
			if err := uc.transactionRepo.Delete(ctx, pair.ID); err != nil {
				uc.logger.Error().Err(err).Uint("pair_id", pair.ID).Msg("failed to delete transfer pair")
				return domain.Transaction{}, err
			}
		}
	}

	// Delete the original row.
	if err := uc.transactionRepo.Delete(ctx, original.ID); err != nil {
		uc.logger.Error().Err(err).Uint("transaction_id", original.ID).Msg("failed to delete original transaction")
		return domain.Transaction{}, err
	}

	if isTransfer {
		debit := domain.Transaction{
			UserID:               target.UserID,
			AccountID:            target.AccountID,
			CategoryID:           target.CategoryID,
			Type:                 domain.TransactionTypeTransfer,
			Amount:               target.Amount,
			Description:          target.Description,
			Date:                 target.Date,
			Month:                target.Month,
			Status:               target.Status,
			DestinationAccountID: target.DestinationAccountID,
		}
		credit := domain.Transaction{
			UserID:               target.UserID,
			AccountID:            *target.DestinationAccountID,
			CategoryID:           target.CategoryID,
			Type:                 domain.TransactionTypeTransfer,
			Amount:               target.Amount,
			Description:          target.Description,
			Date:                 target.Date,
			Month:                target.Month,
			Status:               target.Status,
			DestinationAccountID: &target.AccountID,
		}
		debitTx, _, err := uc.transactionRepo.CreateTransfer(ctx, debit, credit)
		if err != nil {
			uc.logger.Error().Err(err).Msg("failed to recreate transfer pair")
			return domain.Transaction{}, err
		}
		return debitTx, nil
	}

	// Single-row recreate (transfer -> expense/income).
	fresh := domain.Transaction{
		UserID:      target.UserID,
		AccountID:   target.AccountID,
		CategoryID:  target.CategoryID,
		Type:        target.Type,
		Amount:      target.Amount,
		Description: target.Description,
		Date:        target.Date,
		Month:       target.Month,
		Status:      target.Status,
	}
	created, err := uc.transactionRepo.Create(ctx, fresh)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to recreate transaction")
		return domain.Transaction{}, err
	}
	return created, nil
}

// recalculateBudgetSpent atomically recalculates spent for budgets matching the category
func (uc *UpdateUseCase) recalculateBudgetSpent(ctx context.Context, userID, categoryID uint, txDate time.Time) {
	budgets, err := uc.budgetRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to get budgets for spent recalculation")
		return
	}

	for _, b := range budgets {
		if b.CategoryID == categoryID && b.IsActive &&
			!txDate.Before(b.StartDate) && !txDate.After(b.EndDate) {
			if err := uc.budgetRepo.UpdateSpentAtomic(ctx, b.ID, categoryID, b.StartDate, b.EndDate); err != nil {
				uc.logger.Error().Err(err).Uint("budget_id", b.ID).Msg("failed to recalculate budget spent")
			}
		}
	}
}
