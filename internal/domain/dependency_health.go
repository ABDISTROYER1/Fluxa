package domain

import "time"

// DependencyHealthCheck is a sanitized, platform-wide dependency observation.
type DependencyHealthCheck struct {
	Dependency string    `json:"dependency"`
	Status     string    `json:"status"`
	LatencyMS  int64     `json:"latency_ms"`
	CheckedAt  time.Time `json:"checked_at"`
}
