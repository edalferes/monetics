// Package timex provides time parsing/formatting helpers reused across modules.
package timex

import (
	"errors"
	"time"
)

// ErrInvalidDateFormat is returned when a date string cannot be parsed by any supported layout.
var ErrInvalidDateFormat = errors.New("invalid date format: unable to parse date string")

// ParseFlexibleDate tries to parse a date string using multiple formats.
//
// Supported formats:
//   - "2006-01-02"          (ISO date only)
//   - "02/01/2006"          (DD/MM/YYYY, Brazilian format)
//   - time.RFC3339          ("2006-01-02T15:04:05Z07:00")
//   - "2006-01-02T15:04:05" (no timezone)
//   - time.DateOnly         (Go 1.20+ alias for "2006-01-02")
func ParseFlexibleDate(dateStr string) (time.Time, error) {
	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		time.RFC3339,
		"2006-01-02T15:04:05",
		time.DateOnly,
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, dateStr); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, ErrInvalidDateFormat
}
