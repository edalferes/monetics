// Package model contains GORM persistence models for the budget module.
// Models are converted to/from domain entities at the repository boundary.
package model

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// AccountModel is the GORM persistence model for domain.Account.
type AccountModel struct {
	ID          uint               `gorm:"primaryKey"`
	UserID      uint               `gorm:"not null;index:idx_user_accounts;constraint:OnDelete:CASCADE"`
	Name        string             `gorm:"not null;size:100"`
	Type        domain.AccountType `gorm:"not null;size:20"`
	Balance     float64            `gorm:"type:decimal(15,2);default:0"`
	Currency    string             `gorm:"size:3;default:'BRL'"`
	Description string             `gorm:"type:text"`
	IsActive    bool               `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName overrides the default table name.
func (AccountModel) TableName() string { return "budget_accounts" }

// ToDomain converts an AccountModel to a domain.Account.
func (m AccountModel) ToDomain() domain.Account {
	return domain.Account{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Type:        m.Type,
		Balance:     m.Balance,
		Currency:    m.Currency,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// AccountFromDomain creates an AccountModel from a domain.Account.
func AccountFromDomain(a domain.Account) AccountModel {
	return AccountModel{
		ID:          a.ID,
		UserID:      a.UserID,
		Name:        a.Name,
		Type:        a.Type,
		Balance:     a.Balance,
		Currency:    a.Currency,
		Description: a.Description,
		IsActive:    a.IsActive,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// AccountModelsToDomain converts a slice of AccountModel to a slice of domain.Account.
func AccountModelsToDomain(models []AccountModel) []domain.Account {
	out := make([]domain.Account, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out
}
