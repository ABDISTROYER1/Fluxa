# Dependency health and status

The API exposes platform status without tenant authentication:

- `GET /status` combines active incident severity with recent dependency health. A current unhealthy check yields degraded status; if all monitored checks are unhealthy, status is outage. Missing or older than three minutes checks are `unknown` and produce degraded status. Incident records continue to appear in `recent_incidents`.
- `GET /status/dependencies/history` returns observations newest first. `dependency` filters by dependency name, `since` accepts an RFC3339 timestamp, and `limit` defaults to 100 and is capped at 500.

The API process samples PostgreSQL, replica, Redis, Horizon, and worker heartbeat at startup and once per minute. Persisted fields are dependency name, healthy/unhealthy status, latency in milliseconds, and sample time. Probe error messages, credentials, and response bodies are discarded. Rows older than 30 days are pruned daily.

The status history table is created by the standard API database migration. No additional environment variables are required. If the database is unavailable, status reads use the standard API error envelope; `/health/ready` remains the direct readiness signal.