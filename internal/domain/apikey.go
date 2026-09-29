package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	ScopeTransfersRead    = "transfers:read"
	ScopeTransfersWrite   = "transfers:write"
	ScopeWalletsRead      = "wallets:read"
	ScopeWalletsWrite     = "wallets:write"
	ScopeWebhooksRead     = "webhooks:read"
	ScopeWebhooksWrite    = "webhooks:write"
	ScopeKeysRead         = "keys:read"
	ScopeKeysWrite        = "keys:write"
	ScopeAuditRead        = "audit:read"
	ScopeFiatRead         = "fiat:read"
	ScopeFiatWrite        = "fiat:write"
	ScopeComplianceRead   = "compliance:read"
	ScopeComplianceWrite  = "compliance:write"
	ScopeFXRead           = "fx:read"
	ScopeFXWrite          = "fx:write"
	ScopeFeesRead         = "fees:read"
	ScopeWildcard         = "*"
)

var ValidScopes = map[string]bool{
	ScopeTransfersRead:   true,
	ScopeTransfersWrite:  true,
	ScopeWalletsRead:     true,
	ScopeWalletsWrite:    true,
	ScopeWebhooksRead:    true,
	ScopeWebhooksWrite:   true,
	ScopeKeysRead:        true,
	ScopeKeysWrite:       true,
	ScopeAuditRead:       true,
	ScopeFiatRead:        true,
	ScopeFiatWrite:       true,
	ScopeComplianceRead:  true,
	ScopeComplianceWrite: true,
	ScopeFXRead:          true,
	ScopeFXWrite:         true,
	ScopeFeesRead:        true,
	ScopeWildcard:        true,
	"admin":              true,
	"transfers:*":        true,
	"wallets:*":          true,
	"webhooks:*":         true,
	"keys:*":             true,
	"fiat:*":             true,
	"compliance:*":       true,
	"fx:*":               true,
	"fees:*":             true,
	"audit:*":            true,
}

type APIKey struct {
	ID         string
	TenantID   string
	KeyHash    string
	Prefix     string
	Mode       Mode
	Label      *string
	Role       string
	Scopes     []string
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

// HasScope checks if grantedScopes satisfy requiredScope.
// An empty slice of grantedScopes represents unrestricted access (backwards compatibility).
func HasScope(grantedScopes []string, requiredScope string) bool {
	if len(grantedScopes) == 0 {
		return true
	}
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
