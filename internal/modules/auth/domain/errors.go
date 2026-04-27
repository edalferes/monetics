package domain

import "errors"

// Password policy violations. They live here (not in the errors package) so
// the `errors` package can re-export them and the domain remains self-contained.
var (
	ErrPasswordTooShort       = errors.New("password is too short")
	ErrPasswordMissingUpper   = errors.New("password must contain an upper case letter")
	ErrPasswordMissingLower   = errors.New("password must contain a lower case letter")
	ErrPasswordMissingDigit   = errors.New("password must contain a digit")
	ErrPasswordMissingSpecial = errors.New("password must contain a special character")
)
