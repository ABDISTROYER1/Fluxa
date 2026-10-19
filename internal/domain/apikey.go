package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	ScopeTransfersRead            = "transfers:read"
	ScopeTransfersWrite           = "transfers:write"
	ScopeWalletsRead              = "wallets:read"
	ScopeWalletsWrite             = "wallets:write"
	ScopeWebhooksRead             = "webhooks:read"
	ScopeWebhooksWrite            = "webhooks:write"
	ScopeKeysRead                 = "keys:read"
	ScopeKeysWrite                = "keys:write"
	ScopeAuditRead                = "audit:read"
	ScopeFiatRead                 = "fiat:read"
	ScopeFiatWrite                = "fiat:write"
	ScopeComplianceRead           = "compliance:read"
	ScopeComplianceWrite          = "compliance:write"
	ScopeFXRead                   = "fx:read"
	ScopeFXWrite                  = "fx:write"
	ScopeFeesRead                 = "fees:read"
	ScopeBeneficiariesRead        = "beneficiaries:read"
	ScopeBeneficiariesWrite       = "beneficiaries:write"
	ScopeBatchesRead              = "batches:read"
	ScopeBatchesWrite             = "batches:write"
	ScopeReportsRead              = "reports:read"
	ScopeReportsWrite             = "reports:write"
	ScopeWalletBalanceAlertsRead  = "wallet_balance_alerts:read"
	ScopeWalletBalanceAlertsWrite = "wallet_balance_alerts:write"
	ScopeWildcard                 = "*"
)

var ValidScopes = map[string]bool{
	ScopeTransfersRead:            true,
	ScopeTransfersWrite:           true,
	ScopeWalletsRead:              true,
	ScopeWalletsWrite:             true,
	ScopeWebhooksRead:             true,
	ScopeWebhooksWrite:            true,
	ScopeKeysRead:                 true,
	ScopeKeysWrite:                true,
	ScopeAuditRead:                true,
	ScopeFiatRead:                 true,
	ScopeFiatWrite:                true,
	ScopeComplianceRead:           true,
	ScopeComplianceWrite:          true,
	ScopeFXRead:                   true,
	ScopeFXWrite:                  true,
	ScopeFeesRead:                 true,
	ScopeBeneficiariesRead:        true,
	ScopeBeneficiariesWrite:       true,
	ScopeBatchesRead:              true,
	ScopeBatchesWrite:             true,
	ScopeReportsRead:              true,
	ScopeReportsWrite:             true,
	ScopeWalletBalanceAlertsRead:  true,
	ScopeWalletBalanceAlertsWrite: true,
	ScopeWildcard:                 true,
	"admin":                       true,
	"transfers:*":                 true,
	"wallets:*":                   true,
	"webhooks:*":                  true,
	"keys:*":                      true,
	"fiat:*":                      true,
	"compliance:*":                true,
	"fx:*":                        true,
	"fees:*":                      true,
	"beneficiaries:*":             true,
	"batches:*":                   true,
	"reports:*":                   true,
	"wallet_balance_alerts:*":     true,
	"audit:*":                     true,
}

var legacyScopeGrants = map[string][]string{
	ScopeBatchesRead:  {ScopeTransfersRead},
	ScopeBatchesWrite: {ScopeTransfersWrite},
}

type APIKey struct {
	ID                     string
	TenantID               string
	KeyHash                string
	Prefix                 string
	Mode                   Mode
	Label                  *string
	Role                   string
	Scopes                 []string
	LastUsedAt             *time.Time
	RevokedAt              *time.Time
	ExpiresAt              *time.Time
	RotationReminderDays   int
	LastRotationRemindedAt *time.Time
	CreatedAt              time.Time
}

// IsExpired returns true if the key has an expiry set and the given time is after it.
func (k *APIKey) IsExpired(at time.Time) bool {
	if k.ExpiresAt == nil {
		return false
	}
	return at.After(*k.ExpiresAt)
}

// NeedsRotationReminder returns true if an unrevoked, unexpired key is within
// its rotation reminder window and has not yet been reminded within that window.
func (k *APIKey) NeedsRotationReminder(now time.Time) bool {
	if k.RevokedAt != nil || k.ExpiresAt == nil || k.IsExpired(now) {
		return false
	}
	reminderDays := k.RotationReminderDays
	if reminderDays <= 0 {
		reminderDays = 7
	}
	reminderThreshold := k.ExpiresAt.AddDate(0, 0, -reminderDays)
	if now.Before(reminderThreshold) {
		return false
	}
	// If already reminded after the threshold began, no need to spam.
	if k.LastRotationRemindedAt != nil && !k.LastRotationRemindedAt.Before(reminderThreshold) {
		return false
	}
	return true
}

// HasScope checks if grantedScopes satisfy requiredScope.
// An empty slice of grantedScopes represents unrestricted access: keys created
// before scopes existed were migrated with no scopes and stay full-access.
func HasScope(grantedScopes []string, requiredScope string) bool {
	if len(grantedScopes) == 0 {
		return true
	}
	if grantsScope(grantedScopes, requiredScope) {
		return true
	}
	for _, legacy := range legacyScopeGrants[requiredScope] {
		if grantsScope(grantedScopes, legacy) {
			return true
		}
	}
	return false
}

func grantsScope(grantedScopes []string, requiredScope string) bool {
	for _, s := range grantedScopes {
		if s == ScopeWildcard || s == "admin" || s == requiredScope {
			return true
		}
		// Resource wildcard e.g. "transfers:*" matches "transfers:read"
		if strings.HasSuffix(s, ":*") {
			prefix := strings.TrimSuffix(s, "*")
			if strings.HasPrefix(requiredScope, prefix) {
				return true
			}
		}
	}
	return false
}

// ValidateScopes verifies that all provided scopes are recognized.
func ValidateScopes(scopes []string) error {
	for _, s := range scopes {
		if !ValidScopes[s] {
			return fmt.Errorf("invalid scope: %s", s)
		}
	}
	return nil
}
