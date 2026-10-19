package domain

import (
	"strings"
	"time"
)

const (
	DiscrepancyMissingOnChain      = "missing_onchain_transaction"
	DiscrepancyAmountMismatch      = "amount_mismatch"
	DiscrepancyAssetMismatch       = "asset_mismatch"
	DiscrepancyDuplicateSettlement = "duplicate_settlement"
	DiscrepancyStalePending        = "stale_pending_state"
)

func ReconciliationDiscrepancyCategory(entry *AuditLogEntry) string {
	details := strings.ToLower(entry.Details)
	status := strings.ToLower(entry.HorizonStatus)
	if entry.Outcome == AuditNotFound {
		return DiscrepancyMissingOnChain
	}
	if strings.Contains(details, "duplicate") || strings.Contains(status, "duplicate") {
		return DiscrepancyDuplicateSettlement
	}
	if strings.Contains(status, "pending") || strings.Contains(details, "stale pending") {
		return DiscrepancyStalePending
	}
	if !entry.AssetVerified {
		return DiscrepancyAssetMismatch
	}
	return DiscrepancyAmountMismatch
}

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

type ReconciliationDiscrepancy struct {
	ID             string                           `json:"id"`
	TenantID       string                           `json:"tenant_id"`
	TransactionID  string                           `json:"transaction_id"`
	Category       string                           `json:"category"`
	Status         string                           `json:"status"`
	LastAuditLogID *string                          `json:"last_audit_log_id,omitempty"`
	AssignedTo     *string                          `json:"assigned_to,omitempty"`
	ResolutionNote string                           `json:"resolution_note,omitempty"`
	DetectedAt     time.Time                        `json:"detected_at"`
	UpdatedAt      time.Time                        `json:"updated_at"`
	ResolvedAt     *time.Time                       `json:"resolved_at,omitempty"`
	History        []ReconciliationDiscrepancyEvent `json:"history"`
	Amount         string                           `json:"amount,omitempty"`
	Asset          string                           `json:"asset,omitempty"`
	StellarTxHash  string                           `json:"stellar_tx_hash,omitempty"`
	Details        string                           `json:"details,omitempty"`
}

type ReconciliationDiscrepancySummary struct {
	Counts             map[string]int `json:"counts"`
	UnresolvedCount    int            `json:"unresolved_count"`
	OldestUnresolvedAt *time.Time     `json:"oldest_unresolved_at,omitempty"`
}

type ReconciliationDiscrepancyEvent struct {
	ID         string    `json:"id"`
	ActorID    *string   `json:"actor_id,omitempty"`
	Action     string    `json:"action"`
	Note       string    `json:"note,omitempty"`
	AssignedTo *string   `json:"assigned_to,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
