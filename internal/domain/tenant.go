package domain

import (
	"time"
)

// AccountType is the tenant's billing/limit profile.
type AccountType string

const (
	AccountTypeIndividual   AccountType = "individual"
	AccountTypeOrganization AccountType = "organization"
)

const (
	// defaultIndividualWalletLimit and defaultOrganizationWalletLimit apply
	// when a tenant row has no explicit max_wallets.
	defaultIndividualWalletLimit   = 5
	defaultOrganizationWalletLimit = 50
)

type Tenant struct {
	ID                   string      `json:"id" db:"id"`
	Name                 string      `json:"name" db:"name"`
	Email                string      `json:"email,omitempty" db:"email"`
	SigningSecret        string      `json:"signing_secret,omitempty" db:"signing_secret"`
	AccountType          AccountType `json:"account_type" db:"account_type"`
	MaxWallets           int         `json:"max_wallets" db:"max_wallets"`
	MaxTransfersPerMonth *int        `json:"max_transfers_per_month" db:"max_transfers_per_month"`
	MaxWebhooks          int         `json:"max_webhooks" db:"max_webhooks"`
	CreatedAt            time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at" db:"updated_at"`
}

// GetTransferLimit returns the tenant's monthly transfer cap. A nil value
// means the tenant has no configured cap; 0 from this method disables the
// limit check at the call site.
func (t *Tenant) GetTransferLimit() int {
	if t.MaxTransfersPerMonth == nil {
		return 0
	}
	return *t.MaxTransfersPerMonth
}

// GetWalletLimit returns the tenant's wallet cap, falling back to the
// account-type default when max_wallets is unset.
func (t *Tenant) GetWalletLimit() int {
	if t.MaxWallets > 0 {
		return t.MaxWallets
	}
	if t.AccountType == AccountTypeOrganization {
		return defaultOrganizationWalletLimit
	}
	return defaultIndividualWalletLimit
}
