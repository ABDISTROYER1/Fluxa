package domain

import "time"

// ClaimableFilter narrows a claimable balance listing. A zero-valued field is ignored,
// so the zero Filter means "every balance visible to the caller".
type ClaimableFilter struct {
	Status        ClaimableBalanceStatus
	Asset         string
	Claimant      string
	ExpiresBefore *time.Time
	Limit         int
	Offset        int
}
