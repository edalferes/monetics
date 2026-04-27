package domain

import "time"

// AuditLog records security-relevant events for compliance and forensics.
type AuditLog struct {
	ID        uint
	UserID    *uint
	Username  string
	Action    string
	Status    string
	IP        string
	Details   string
	CreatedAt time.Time
}
