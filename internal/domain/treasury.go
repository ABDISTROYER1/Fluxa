package domain

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// TreasuryConfig holds the per-asset sweep policy stored in treasury_config.
type TreasuryConfig struct {
	Asset              string
	SweepThreshold     decimal.Decimal
	MinOperatingBuffer decimal.Decimal
	ColdStorageAddress string
	AutoSweepEnabled   bool
	UpdatedAt          time.Time
}

// TreasurySweepLog is one row of the sweep audit trail, written on every sweep
// attempt — including "zero sweeps" where nothing was moved.
type TreasurySweepLog struct {
	ID          string
	Asset       string
	Amount      decimal.Decimal
	Destination string
	TxHash      string
	TriggeredBy string // "auto" | "manual"
	SweptAt     time.Time
}

const (
	TriggeredByAuto   = "auto"
	TriggeredByManual = "manual"
)

// IdempotencyRecord is a persisted idempotency request and, once complete, its exact
// response. ID and LeaseToken fence stale owners from overwriting a recovered
// record.
type IdempotencyRecord struct {
	ID              string
	OrgID           string
	Mode            Mode
	Key             string
	RequestHash     string
	Status          string
	LeaseToken      string
	LeaseExpiresAt  time.Time
	ExpiresAt       time.Time
	ResponseStatus  int
	ResponseHeaders http.Header
	ResponseBody    []byte
}
