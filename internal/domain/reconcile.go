package domain

import "time"

type AuditOutcome string

const (
	AuditOK       AuditOutcome = "ok"
	AuditMismatch AuditOutcome = "mismatch"
	AuditNotFound AuditOutcome = "not_found"
)

// AuditLogEntry records the outcome of a single transaction reconciliation check.
type AuditLogEntry struct {
	ID             string
	TxID           string
	StellarHash    string
	CheckedAt      time.Time
	HorizonStatus  string
	AmountVerified bool
	AssetVerified  bool
	FeeVerified    bool
	Outcome        AuditOutcome
	Details        string
}

// DailySummaryRow holds aggregated reconciliation counts for a single day.
type DailySummaryRow struct {
	Date          string
	OKCount       int
	MismatchCount int
	NotFoundCount int
}

// ReconciliationRun records the outcome of a single reconciliation pass.
type ReconciliationRun struct {
	ID                 string
	StartedAt          time.Time
	CompletedAt        time.Time
	TxsChecked         int
	DiscrepanciesFound int
	CorrectionsMade    int
}
