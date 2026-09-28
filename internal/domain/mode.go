package domain

import "fmt"

// Mode identifies an isolated Stellar environment for one tenant.
type Mode string

const (
	ModeLive Mode = "live"
	ModeTest Mode = "test"
)

func (m Mode) Valid() bool {
	return m == ModeLive || m == ModeTest
}

func ParseMode(value string) (Mode, error) {
	mode := Mode(value)
	if !mode.Valid() {
		return "", fmt.Errorf("invalid environment mode %q (want live or test)", value)
	}
	return mode, nil
}
