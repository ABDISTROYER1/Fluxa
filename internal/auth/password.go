package auth

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	// MinPasswordLength is the minimum allowed password length in characters.
	MinPasswordLength = 8

	// MaxPasswordLength is the maximum allowed password length in bytes.
	// bcrypt operates on a maximum of 72 bytes; inputs beyond 72 bytes
	// are either rejected or silently truncated by bcrypt implementations.
	MaxPasswordLength = 72
)

var (
	// ErrPasswordTooShort is returned when a password has fewer than 8 characters.
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters long", MinPasswordLength)

	// ErrPasswordTooLong is returned when a password exceeds bcrypt's 72-byte limit.
	ErrPasswordTooLong = fmt.Errorf("password must not exceed %d bytes", MaxPasswordLength)
)

// ValidatePassword checks password policy compliance.
//
// Policy Decision (NIST SP 800-63B):
//   - Length: Minimum 8 characters to ensure adequate baseline entropy.
//   - Maximum Length: Exactly 72 bytes to strictly match bcrypt's input limit.
//     Passwords longer than 72 bytes are rejected with ErrPasswordTooLong rather
//     than silently truncated, ensuring users know that trailing characters
//     cannot contribute entropy.
//   - Complexity: Arbitrary composition rules (mandatory digits, uppercase, special
//     characters) are deliberately omitted per NIST SP 800-63B recommendations,
//     as they encourage predictable patterns (e.g. "Password1!") and degrade usability
//     without improving security.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len([]byte(password)) > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

// IsPasswordValidationError returns true if err is a password policy validation error.
func IsPasswordValidationError(err error) bool {
	return errors.Is(err, ErrPasswordTooShort) || errors.Is(err, ErrPasswordTooLong)
}
