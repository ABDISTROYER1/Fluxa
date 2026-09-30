package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClaimableBalanceStatus string

const (
	ClaimableBalanceStatusPending ClaimableBalanceStatus = "pending"
	ClaimableBalanceStatusClaimed ClaimableBalanceStatus = "claimed"
	ClaimableBalanceStatusRevoked ClaimableBalanceStatus = "revoked"
	ClaimableBalanceStatusExpired ClaimableBalanceStatus = "expired"
)

type PredicateType string

const (
	PredicateUnconditional      PredicateType = "unconditional"
	PredicateBeforeAbsoluteTime PredicateType = "before_absolute_time"
	PredicateAfterAbsoluteTime  PredicateType = "after_absolute_time"
	PredicateBeforeRelativeTime PredicateType = "before_relative_time"
	PredicateNot                PredicateType = "not"
	PredicateAnd                PredicateType = "and"
	PredicateOr                 PredicateType = "or"
)

type ClaimPredicate struct {
	Type       PredicateType    `json:"type"`
	Timestamp  int64            `json:"timestamp,omitempty"`
	Seconds    int64            `json:"seconds,omitempty"`
	Predicates []ClaimPredicate `json:"predicates,omitempty"`
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
