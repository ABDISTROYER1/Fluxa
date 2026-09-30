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

type ClaimPredicateType string

const (
	PredicateUnconditional      ClaimPredicateType = "unconditional"
	PredicateBeforeAbsoluteTime ClaimPredicateType = "before_absolute_time"
	PredicateAfterAbsoluteTime  ClaimPredicateType = "after_absolute_time"
	PredicateBeforeRelativeTime ClaimPredicateType = "before_relative_time"
	PredicateNot                ClaimPredicateType = "not"
	PredicateAnd                ClaimPredicateType = "and"
	PredicateOr                 ClaimPredicateType = "or"
)

type ClaimPredicate struct {
	Type       ClaimPredicateType `json:"type"`
	Timestamp  int64              `json:"timestamp,omitempty"`
	Seconds    int64              `json:"seconds,omitempty"`
	Predicates []ClaimPredicate   `json:"predicates,omitempty"`
}

type Claimant struct {
	Account   string          `json:"account"`
	Predicate *ClaimPredicate `json:"predicate,omitempty"`
}

type ClaimableBalance struct {
	ID             string
	TenantID       *string
	Asset          string
	Amount         decimal.Decimal
	Claimants      []Claimant
	Sponsor        string
	Status         ClaimableBalanceStatus
	RevokeOnExpiry bool
	CreatedAt      time.Time
	ExpiresAt      *time.Time
	ClaimedAt      *time.Time
	ClaimedBy      string
}

// SourceWallet is a Fluxa-custodied Stellar account usable as the funding or
// sponsor account for a claimable balance.
type SourceWallet struct {
	ID string
	// PublicKey is the Stellar account ID (G...).
	PublicKey string
	// EncryptedSecret is the hex-encoded AES-GCM envelope of the account seed,
	// exactly as stored on Wallet.
	EncryptedSecret string
}
