package model

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
)

// CategoryModel is the GORM persistence model for domain.Category.
type CategoryModel struct {
	ID          uint                `gorm:"primaryKey"`
	UserID      uint                `gorm:"not null;index:idx_user_categories;constraint:OnDelete:CASCADE"`
	Name        string              `gorm:"not null;size:100"`
	Type        domain.CategoryType `gorm:"not null;size:20"`
	Icon        string              `gorm:"size:50"`
	Color       string              `gorm:"size:20"`
	Description string              `gorm:"type:text"`
	IsActive    bool                `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CategoryModel) TableName() string { return "budget_categories" }

func (m CategoryModel) ToDomain() domain.Category {
	return domain.Category{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Type:        m.Type,
		Icon:        m.Icon,
		Color:       m.Color,
		Description: m.Description,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func CategoryFromDomain(c domain.Category) CategoryModel {
	return CategoryModel{
		ID:          c.ID,
		UserID:      c.UserID,
		Name:        c.Name,
		Type:        c.Type,
		Icon:        c.Icon,
		Color:       c.Color,
		Description: c.Description,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func CategoryModelsToDomain(models []CategoryModel) []domain.Category {
	out := make([]domain.Category, len(models))
	for i, m := range models {
		out[i] = m.ToDomain()
	}
	return out
}
