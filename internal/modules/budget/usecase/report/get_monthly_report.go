package report

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
)

// GetMonthlyReportUseCase generates monthly financial report
type GetMonthlyReportUseCase struct {
	transactionRepo interfaces.TransactionRepository
	budgetRepo      interfaces.BudgetRepository
	categoryRepo    interfaces.CategoryRepository
}

// NewGetMonthlyReportUseCase creates a new use case instance
func NewGetMonthlyReportUseCase(
	transactionRepo interfaces.TransactionRepository,
	budgetRepo interfaces.BudgetRepository,
	categoryRepo interfaces.CategoryRepository,
) *GetMonthlyReportUseCase {
	return &GetMonthlyReportUseCase{
		transactionRepo: transactionRepo,
		budgetRepo:      budgetRepo,
		categoryRepo:    categoryRepo,
	}
}

// Execute generates a monthly report
func (uc *GetMonthlyReportUseCase) Execute(ctx context.Context, userID uint, year int, month time.Month) (*MonthlyReport, error) {
	// Calculate date range for the month
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get all transactions for the month
	transactions, err := uc.transactionRepo.GetByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get all categories
	categories, err := uc.categoryRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate totals
	var totalIncome, totalExpense float64
	categoryTotals := make(map[uint]CategoryTotal)

	for _, tx := range transactions {
		if tx.Status != domain.TransactionStatusCompleted {
			continue
		}

		// Find category
		var categoryName string
		var categoryType domain.CategoryType
		for _, cat := range categories {
			if cat.ID == tx.CategoryID {
				categoryName = cat.Name
				categoryType = cat.Type
				break
			}
		}

		switch tx.Type {
		case domain.TransactionTypeIncome:
			totalIncome += tx.Amount
			if total, exists := categoryTotals[tx.CategoryID]; exists {
				total.Amount += tx.Amount
				categoryTotals[tx.CategoryID] = total
			} else {
				categoryTotals[tx.CategoryID] = CategoryTotal{
					CategoryID:   tx.CategoryID,
					CategoryName: categoryName,
					CategoryType: categoryType,
					Amount:       tx.Amount,
				}
			}
		case domain.TransactionTypeExpense:
			totalExpense += tx.Amount
			if total, exists := categoryTotals[tx.CategoryID]; exists {
				total.Amount += tx.Amount
				categoryTotals[tx.CategoryID] = total
			} else {
				categoryTotals[tx.CategoryID] = CategoryTotal{
					CategoryID:   tx.CategoryID,
					CategoryName: categoryName,
					CategoryType: categoryType,
					Amount:       tx.Amount,
				}
			}
		case domain.TransactionTypeTransfer:
			// Transfer fees count as expense
			if tx.TransferFee != nil && *tx.TransferFee > 0 {
				totalExpense += *tx.TransferFee
			}
		}
	}

	// Convert map to slice
	categoryTotalsList := make([]CategoryTotal, 0, len(categoryTotals))
	for _, total := range categoryTotals {
		categoryTotalsList = append(categoryTotalsList, total)
	}

	return &MonthlyReport{
		Year:           year,
		Month:          month,
		TotalIncome:    totalIncome,
		TotalExpense:   totalExpense,
		Balance:        totalIncome - totalExpense,
		CategoryTotals: categoryTotalsList,
		Transactions:   transactions,
	}, nil
}

// MonthlyReport represents a monthly financial report
type MonthlyReport struct {
	Year           int                  `json:"year"`
	Month          time.Month           `json:"month"`
	TotalIncome    float64              `json:"total_income"`
	TotalExpense   float64              `json:"total_expense"`
	Balance        float64              `json:"balance"`
	CategoryTotals []CategoryTotal      `json:"category_totals"`
	Transactions   []domain.Transaction `json:"transactions,omitempty"`
}

// CategoryTotal represents spending/income by category
type CategoryTotal struct {
	CategoryID   uint                `json:"category_id"`
	CategoryName string              `json:"category_name"`
	CategoryType domain.CategoryType `json:"category_type"`
	Amount       float64             `json:"amount"`
}
