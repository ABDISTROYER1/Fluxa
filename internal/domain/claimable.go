package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClaimableBalanceStatus string

const (
	ClaimableBalanceStatusPending ClaimableBalanceStatus = "pending"
	ClaimableBalanceStatusClaimed ClaimableBalanceStatus = "claimed"
	ClaimableBalanceStatusExpired ClaimableBalanceStatus = "expired"
	ClaimableBalanceStatusRevoked ClaimableBalanceStatus = "revoked"
)

type ClaimableBalance struct {
	ID               string
	TenantID         string
	Mode             Mode
	BalanceID        string
	SponsorWalletID  string
	ClaimantAddress  string
	Asset            string
	Amount           decimal.Decimal
	Status           ClaimableBalanceStatus
	PredicateType    string
	PredicateExpiry  *time.Time
	ExpiresAt        *time.Time
	Revocable        bool
	ClaimedAt        *time.Time
	ClaimedByAddress *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SourceWallet struct {
	ID        string
	PublicKey string
}
