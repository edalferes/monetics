package report

import (
	"context"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// GetAccountBalanceUseCase handles getting account with calculated balance
type GetAccountBalanceUseCase struct {
	accountRepo     interfaces.AccountRepository
	transactionRepo interfaces.TransactionRepository
}

// NewGetAccountBalanceUseCase creates a new use case instance
func NewGetAccountBalanceUseCase(
	accountRepo interfaces.AccountRepository,
	transactionRepo interfaces.TransactionRepository,
) *GetAccountBalanceUseCase {
	return &GetAccountBalanceUseCase{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

// AccountBalanceOutput represents the account with calculated balance
type AccountBalanceOutput struct {
	Account        domain.Account `json:"account"`
	CurrentBalance float64        `json:"current_balance"`
	TotalIncome    float64        `json:"total_income"`
	TotalExpense   float64        `json:"total_expense"`
	TotalTransfers float64        `json:"total_transfers"`
}

// Execute gets account and calculates its current balance
func (uc *GetAccountBalanceUseCase) Execute(ctx context.Context, userID uint, accountID uint) (AccountBalanceOutput, error) {
	// Get account
	account, err := uc.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return AccountBalanceOutput{}, errors.ErrAccountNotFound
	}

	// Verify ownership
	if account.UserID != userID {
		return AccountBalanceOutput{}, errors.ErrUnauthorizedAccess
	}

	// Get all transactions for this account
	transactions, err := uc.transactionRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return AccountBalanceOutput{}, err
	}

	// Calculate balance.
	// Transfers are persisted as a debit/credit pair (RN-T3): the debit row has
	// AccountID=source and DestinationAccountID set; the credit row has
	// AccountID=destination and ParentID set. Each row affects only its own
	// AccountID, so the loop simply applies the sign based on the row kind.
	var totalIncome, totalExpense, totalTransfers float64
	for _, tx := range transactions {
		switch tx.Type {
		case domain.TransactionTypeIncome:
			totalIncome += tx.Amount
		case domain.TransactionTypeExpense:
			totalExpense += tx.Amount
		case domain.TransactionTypeTransfer:
			if tx.ParentID != nil {
				// Credit row: incoming transfer.
				totalTransfers += tx.Amount
			} else {
				// Debit row: outgoing transfer (plus optional fee).
				totalTransfers -= tx.Amount
				if tx.TransferFee != nil {
					totalTransfers -= *tx.TransferFee
				}
			}
		}
	}

	// Current balance = initial balance + income - expense + transfers
	currentBalance := account.Balance + totalIncome - totalExpense + totalTransfers

	return AccountBalanceOutput{
		Account:        account,
		CurrentBalance: currentBalance,
		TotalIncome:    totalIncome,
		TotalExpense:   totalExpense,
		TotalTransfers: totalTransfers,
	}, nil
}
