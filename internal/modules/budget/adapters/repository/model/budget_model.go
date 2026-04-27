package model

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// BudgetModel is the GORM persistence model for domain.Budget.
type BudgetModel struct {
	ID          uint                `gorm:"primaryKey"`
	UserID      uint                `gorm:"not null;index:idx_user_budgets;constraint:OnDelete:CASCADE"`
	CategoryID  uint                `gorm:"not null;index:idx_category_budgets"`
	Name        string              `gorm:"not null;size:200"`
	Amount      float64             `gorm:"type:decimal(15,2);not null"`
	Spent       float64             `gorm:"type:decimal(15,2);default:0"`
	Period      domain.BudgetPeriod `gorm:"not null;size:20"`
	StartDate   time.Time           `gorm:"not null;index:idx_budget_period"`
	EndDate     time.Time           `gorm:"not null;index:idx_budget_period"`
	AlertAt     *float64            `gorm:"type:decimal(5,2)"`
	IsActive    bool                `gorm:"default:true"`
	Description string              `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (BudgetModel) TableName() string { return "budget_budgets" }

func (m BudgetModel) ToDomain() domain.Budget {
	return domain.Budget{
		ID:          m.ID,
		UserID:      m.UserID,
		CategoryID:  m.CategoryID,
		Name:        m.Name,
		Amount:      m.Amount,
		Spent:       m.Spent,
		Period:      m.Period,
		StartDate:   m.StartDate,
		EndDate:     m.EndDate,
		AlertAt:     m.AlertAt,
		IsActive:    m.IsActive,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func BudgetFromDomain(b domain.Budget) BudgetModel {
	return BudgetModel{
		ID:          b.ID,
		UserID:      b.UserID,
		CategoryID:  b.CategoryID,
		Name:        b.Name,
		Amount:      b.Amount,
		Spent:       b.Spent,
		Period:      b.Period,
		StartDate:   b.StartDate,
		EndDate:     b.EndDate,
		AlertAt:     b.AlertAt,
		IsActive:    b.IsActive,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func BudgetModelsToDomain(models []BudgetModel) []domain.Budget {
	out := make([]domain.Budget, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out
}
