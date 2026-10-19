package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type RateResponse struct {
	Rate          decimal.Decimal `json:"rate"`
	MidMarketRate decimal.Decimal `json:"mid_market_rate"`
	SpreadBps     int             `json:"spread_bps"`
	Provider      string          `json:"provider"`
	CachedAt      time.Time       `json:"cached_at"`
	Stale         bool            `json:"stale"`
}
