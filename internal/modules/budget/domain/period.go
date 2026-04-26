package domain

import (
	"errors"
	"time"
)

// ErrInvalidPeriod is returned when a period would have an inverted or
// zero-length range.
var ErrInvalidPeriod = errors.New("invalid period: start must be strictly before end")

// Period represents a closed time interval [Start, End).
//
// It pairs raw start/end timestamps with a logical BudgetPeriod (daily,
// weekly, monthly...) so business rules can compare consistent units.
type Period struct {
	Start time.Time
	End   time.Time
	Type  BudgetPeriod
}

// NewPeriod builds a Period and rejects invalid ranges.
func NewPeriod(start, end time.Time, periodType BudgetPeriod) (Period, error) {
	if !start.Before(end) {
		return Period{}, ErrInvalidPeriod
	}
	return Period{Start: start, End: end, Type: periodType}, nil
}

// Contains reports whether the given timestamp falls within the period.
func (p Period) Contains(t time.Time) bool {
	return !t.Before(p.Start) && !t.After(p.End)
}

// Overlaps reports whether two periods share at least one instant.
func (p Period) Overlaps(other Period) bool {
	return p.Start.Before(other.End) && other.Start.Before(p.End)
}

// Duration returns the length of the period.
func (p Period) Duration() time.Duration { return p.End.Sub(p.Start) }
