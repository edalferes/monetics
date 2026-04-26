// Package domain — password policy lives here so use cases stay framework-free.
package domain

import "unicode"

// PasswordPolicy enforces password strength rules.
//
// The default policy requires:
//   - at least 8 characters
//   - at least one upper case letter
//   - at least one lower case letter
//   - at least one digit
//
// The policy is a value object: callers can build a stricter one if needed.
type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPasswordPolicy returns the project-wide default policy.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:    8,
		RequireUpper: true,
		RequireLower: true,
		RequireDigit: true,
	}
}

// Validate reports whether the given plaintext password satisfies the policy.
// It returns nil on success or a domain-level error describing the first
// violated rule.
func (p PasswordPolicy) Validate(password string) error {
	if len(password) < p.MinLength {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return ErrPasswordMissingUpper
	}
	if p.RequireLower && !hasLower {
		return ErrPasswordMissingLower
	}
	if p.RequireDigit && !hasDigit {
		return ErrPasswordMissingDigit
	}
	if p.RequireSpecial && !hasSpecial {
		return ErrPasswordMissingSpecial
	}
	return nil
}
