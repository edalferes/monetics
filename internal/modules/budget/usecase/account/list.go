package account

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
)

// ListUseCase handles listing user accounts
type ListUseCase struct {
	accountRepo     interfaces.AccountRepository
	transactionRepo interfaces.TransactionRepository
	logger          logger.Logger
}

// NewListUseCase creates a new use case instance
func NewListUseCase(accountRepo interfaces.AccountRepository, transactionRepo interfaces.TransactionRepository, log logger.Logger) *ListUseCase {
	return &ListUseCase{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		logger:          log.With().Str("usecase", "account.list").Logger(),
	}
}

// Execute lists all accounts for a user with calculated balances
func (uc *ListUseCase) Execute(ctx context.Context, userID uint) ([]domain.Account, error) {
	uc.logger.Debug().Uint("user_id", userID).Msg("listing accounts")

	accounts, err := uc.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", userID).Msg("failed to list accounts")
		return nil, err
	}

	// Calculate current balance for each account based on transactions
	for i := range accounts {
		transactions, err := uc.transactionRepo.GetByAccountID(ctx, accounts[i].ID)
		if err != nil {
			uc.logger.Error().Err(err).Uint("account_id", accounts[i].ID).Msg("failed to get transactions for balance calculation")
			continue
		}

		var totalIncome, totalExpense, totalTransfers float64
		for _, tx := range transactions {
			switch tx.Type {
			case domain.TransactionTypeIncome:
				totalIncome += tx.Amount
			case domain.TransactionTypeExpense:
				totalExpense += tx.Amount
			case domain.TransactionTypeTransfer:
				if tx.AccountID == accounts[i].ID {
					totalTransfers -= tx.Amount
					if tx.TransferFee != nil {
						totalTransfers -= *tx.TransferFee
					}
				}
				if tx.DestinationAccountID != nil && *tx.DestinationAccountID == accounts[i].ID {
					totalTransfers += tx.Amount
				}
			}
		}

		accounts[i].Balance = accounts[i].Balance + totalIncome - totalExpense + totalTransfers
	}

	uc.logger.Info().Uint("user_id", userID).Int("count", len(accounts)).Msg("accounts listed successfully")
	return accounts, nil
}
