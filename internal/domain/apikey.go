package domain

import "time"

type APIKey struct {
	ID         string
	TenantID   string
	KeyHash    string
	Prefix     string
	Label      *string
	Role       string
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}
