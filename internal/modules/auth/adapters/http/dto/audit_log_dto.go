package dto

import (
	"time"

	"github.com/edalferes/monetics/internal/modules/auth/domain"
)

// AuditLogResponse represents an audit log entry in API responses.
type AuditLogResponse struct {
	ID        uint      `json:"id"`
	UserID    *uint     `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	IP        string    `json:"ip"`
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToAuditLogResponse converts a domain.AuditLog to its API representation.
func ToAuditLogResponse(l domain.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		Username:  l.Username,
		Action:    l.Action,
		Status:    l.Status,
		IP:        l.IP,
		Details:   l.Details,
		CreatedAt: l.CreatedAt,
	}
}

// ToAuditLogResponseList converts a slice of domain.AuditLog.
func ToAuditLogResponseList(logs []domain.AuditLog) []AuditLogResponse {
	out := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		out[i] = ToAuditLogResponse(l)
	}
	return out
}
