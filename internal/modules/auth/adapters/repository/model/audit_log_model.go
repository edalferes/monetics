package model

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/auth/domain"
)

// AuditLogModel is the GORM persistence model for domain.AuditLog.
//
// Table: audit_logs.
type AuditLogModel struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    *uint     `gorm:"index"`
	Username  string    `gorm:"not null"`
	Action    string    `gorm:"not null;index"`
	Status    string    `gorm:"not null;index"`
	IP        string    `gorm:"size:45"`
	Details   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"index"`
}

func (AuditLogModel) TableName() string { return "audit_logs" }

func (m AuditLogModel) ToDomain() domain.AuditLog {
	return domain.AuditLog{
		ID:        m.ID,
		UserID:    m.UserID,
		Username:  m.Username,
		Action:    m.Action,
		Status:    m.Status,
		IP:        m.IP,
		Details:   m.Details,
		CreatedAt: m.CreatedAt,
	}
}

func AuditLogFromDomain(a *domain.AuditLog) *AuditLogModel {
	if a == nil {
		return nil
	}
	return &AuditLogModel{
		ID:        a.ID,
		UserID:    a.UserID,
		Username:  a.Username,
		Action:    a.Action,
		Status:    a.Status,
		IP:        a.IP,
		Details:   a.Details,
		CreatedAt: a.CreatedAt,
	}
}

func AuditLogModelsToDomain(ms []AuditLogModel) []domain.AuditLog {
	out := make([]domain.AuditLog, len(ms))
	for i, m := range ms {
		out[i] = m.ToDomain()
	}
	return out
}
