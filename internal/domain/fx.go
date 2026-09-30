package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Quote is a priced, time-limited conversion offer identified by a unique token.
type Quote struct {
	ID                    string
	OrgID                 string
	FromAsset             string
	ToAsset               string
	FromAmount            decimal.Decimal
	ToAmount              decimal.Decimal
	Rate                  decimal.Decimal
	Fee                   decimal.Decimal
	ExpiresAt             time.Time
	Used                  bool
	FromRequiresTrustline bool
	ToRequiresTrustline   bool
}

// AlertDirection specifies whether the alert fires when the rate crosses above or below the target.
type AlertDirection string

const (
	AlertDirectionAtOrAbove AlertDirection = "at_or_above"
	AlertDirectionAtOrBelow AlertDirection = "at_or_below"
)

// RateAlert represents a tenant-configured threshold notification for an FX pair.
type RateAlert struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	FromAsset      string          `json:"from_asset"`
	ToAsset        string          `json:"to_asset"`
	TargetRate     decimal.Decimal `json:"target_rate"`
	Direction      AlertDirection  `json:"direction"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	LastFiredAt    *time.Time      `json:"last_fired_at,omitempty"`
	LastEvalRate   *decimal.Decimal `json:"last_eval_rate,omitempty"`
	Active         bool            `json:"active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// RateLock holds a quoted rate for a bounded window, consumed by exactly one conversion.
type RateLock struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	FromAsset   string          `json:"from_asset"`
	ToAsset     string          `json:"to_asset"`
	LockedRate  decimal.Decimal `json:"locked_rate"`
	Amount      decimal.Decimal `json:"amount"`
	ExpiresAt   time.Time       `json:"expires_at"`
	ConsumedAt  *time.Time      `json:"consumed_at,omitempty"`
	ConsumedBy  *string         `json:"consumed_by,omitempty"` // conversion ID
	CreatedAt   time.Time       `json:"created_at"`
}
