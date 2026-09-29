package domain

import "time"

// BeneficiaryStatus describes whether a beneficiary can receive transfers.
type BeneficiaryStatus string

const (
	BeneficiaryPending BeneficiaryStatus = "pending"
	BeneficiaryActive  BeneficiaryStatus = "active"
	BeneficiaryRevoked BeneficiaryStatus = "revoked"
)

type Beneficiary struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"-"`
	Mode          Mode              `json:"mode"`
	Account       string            `json:"account"`
	Label         string            `json:"label,omitempty"`
	Status        BeneficiaryStatus `json:"status"`
	CooldownUntil time.Time         `json:"cooldown_until"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
