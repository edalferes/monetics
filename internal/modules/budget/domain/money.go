package domain

import (
	"errors"
	"fmt"
	"math"
)

// Money is a value object representing an amount in a single currency.
//
// The amount is stored as integer minor units (e.g. cents) to avoid the
// rounding errors inherent to float64. Construct via NewMoney or
// NewMoneyFromFloat; never zero-value the struct directly.
type Money struct {
	amountMinor int64
	currency    string
}

// ErrInvalidCurrency is returned when a money operation is attempted with a
// missing or mismatched currency.
var ErrInvalidCurrency = errors.New("invalid or mismatched currency")

// NewMoney builds a Money value from a minor-unit amount (e.g. cents).
func NewMoney(amountMinor int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amountMinor: amountMinor, currency: currency}, nil
}

// NewMoneyFromFloat builds a Money value from a float using bank-style
// rounding to the nearest minor unit. Use only at boundaries (parsing user
// input, importing CSVs); never inside business logic.
func NewMoneyFromFloat(amount float64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}
	cents := int64(math.Round(amount * 100))
	return Money{amountMinor: cents, currency: currency}, nil
}

// AmountMinor returns the raw minor-unit value (e.g. cents).
func (m Money) AmountMinor() int64 { return m.amountMinor }

// Currency returns the ISO currency code.
func (m Money) Currency() string { return m.currency }

// Float returns the amount as a float64 (lossy; use for display only).
func (m Money) Float() float64 { return float64(m.amountMinor) / 100 }

// Add returns the sum of two Money values; both must share the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amountMinor: m.amountMinor + other.amountMinor, currency: m.currency}, nil
}

// Sub returns m minus other; both must share the same currency.
func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrInvalidCurrency
	}
	return Money{amountMinor: m.amountMinor - other.amountMinor, currency: m.currency}, nil
}

// IsPositive reports whether the amount is strictly greater than zero.
func (m Money) IsPositive() bool { return m.amountMinor > 0 }

// IsNegative reports whether the amount is strictly less than zero.
func (m Money) IsNegative() bool { return m.amountMinor < 0 }

// String formats the value as "<currency> <amount.cc>".
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, m.Float())
}
