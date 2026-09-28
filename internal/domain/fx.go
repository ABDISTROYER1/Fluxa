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
