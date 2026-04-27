package model

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// TransactionModel is the GORM persistence model for domain.Transaction.
type TransactionModel struct {
	ID                   uint                     `gorm:"primaryKey"`
	UserID               uint                     `gorm:"not null;index:idx_user_transactions;constraint:OnDelete:CASCADE"`
	AccountID            uint                     `gorm:"not null;index:idx_account_transactions"`
	CategoryID           uint                     `gorm:"not null;index:idx_category_transactions"`
	Type                 domain.TransactionType   `gorm:"not null;size:20"`
	Amount               float64                  `gorm:"type:decimal(15,2);not null"`
	Description          string                   `gorm:"type:text"`
	Date                 time.Time                `gorm:"not null;index:idx_transaction_date"`
	Month                string                   `gorm:"size:7;index:idx_transaction_month"`
	Status               domain.TransactionStatus `gorm:"not null;size:20;default:'completed'"`
	DestinationAccountID *uint                    `gorm:"index:idx_destination_account"`
	TransferFee          *float64                 `gorm:"type:decimal(15,2)"`
	IsRecurring          bool                     `gorm:"default:false"`
	RecurrenceRule       string                   `gorm:"size:50"`
	RecurrenceEnd        *time.Time
	ParentID             *uint    `gorm:"index:idx_parent_transaction"`
	Tags                 []string `gorm:"type:text;serializer:json"`
	Attachments          []string `gorm:"type:text;serializer:json"`
	CreatedAt            time.Time
	UpdatedAt            time.Time

	// Relations preloaded by repository.
	Account  *AccountModel  `gorm:"foreignKey:AccountID"`
	Category *CategoryModel `gorm:"foreignKey:CategoryID"`
}

func (TransactionModel) TableName() string { return "budget_transactions" }

func (m TransactionModel) ToDomain() domain.Transaction {
	t := domain.Transaction{
		ID:                   m.ID,
		UserID:               m.UserID,
		AccountID:            m.AccountID,
		CategoryID:           m.CategoryID,
		Type:                 m.Type,
		Amount:               m.Amount,
		Description:          m.Description,
		Date:                 m.Date,
		Month:                m.Month,
		Status:               m.Status,
		DestinationAccountID: m.DestinationAccountID,
		TransferFee:          m.TransferFee,
		IsRecurring:          m.IsRecurring,
		RecurrenceRule:       m.RecurrenceRule,
		RecurrenceEnd:        m.RecurrenceEnd,
		ParentID:             m.ParentID,
		Tags:                 m.Tags,
		Attachments:          m.Attachments,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
	if m.Account != nil {
		acc := m.Account.ToDomain()
		t.Account = &acc
	}
	if m.Category != nil {
		cat := m.Category.ToDomain()
		t.Category = &cat
	}
	return t
}

func TransactionFromDomain(t domain.Transaction) TransactionModel {
	return TransactionModel{
		ID:                   t.ID,
		UserID:               t.UserID,
		AccountID:            t.AccountID,
		CategoryID:           t.CategoryID,
		Type:                 t.Type,
		Amount:               t.Amount,
		Description:          t.Description,
		Date:                 t.Date,
		Month:                t.Month,
		Status:               t.Status,
		DestinationAccountID: t.DestinationAccountID,
		TransferFee:          t.TransferFee,
		IsRecurring:          t.IsRecurring,
		RecurrenceRule:       t.RecurrenceRule,
		RecurrenceEnd:        t.RecurrenceEnd,
		ParentID:             t.ParentID,
		Tags:                 t.Tags,
		Attachments:          t.Attachments,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

func TransactionModelsToDomain(models []TransactionModel) []domain.Transaction {
	out := make([]domain.Transaction, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out
}
