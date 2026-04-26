package budget

import (
	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository/model"
)

// Entities returns all GORM persistence models for the budget module.
// Used by the database layer for AutoMigrate.
func Entities() []interface{} {
	return []interface{}{
		&model.AccountModel{},
		&model.CategoryModel{},
		&model.TransactionModel{},
		&model.BudgetModel{},
	}
}
