package budget

import "github.com/edalferes/monetics/internal/modules/budget/domain"

func isValidBudgetPeriod(period domain.BudgetPeriod) bool {
	switch period {
	case domain.BudgetPeriodDaily,
		domain.BudgetPeriodWeekly,
		domain.BudgetPeriodMonthly,
		domain.BudgetPeriodQuarterly,
		domain.BudgetPeriodYearly,
		domain.BudgetPeriodCustom:
		return true
	default:
		return false
	}
}
