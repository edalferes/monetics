package helpers

import (
	"errors"
	"time"
)

// ParseFlexibleDate tries to parse a date string using multiple formats
// Accepts: "2006-01-02", "2006-01-02T15:04:05Z07:00", etc
//
// This helper is useful for handling date inputs from various sources
// (API requests, CSV imports, etc.) that may use different date formats.
//
// Supported formats:
//   - "2006-01-02" (date only)
//   - "02/01/2006" (DD/MM/YYYY - Brazilian format)
//   - "2006-01-02T15:04:05Z" (RFC3339)
//   - "2006-01-02T15:04:05" (without timezone)
//   - "2006-01-02T15:04:05-07:00" (with timezone offset)
func ParseFlexibleDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",          // Date only (ISO)
		"02/01/2006",          // DD/MM/YYYY (Brazilian format)
		time.RFC3339,          // Full RFC3339
		"2006-01-02T15:04:05", // Without timezone
		time.DateOnly,         // Go 1.20+ date format
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("invalid date format: unable to parse date string")
}
