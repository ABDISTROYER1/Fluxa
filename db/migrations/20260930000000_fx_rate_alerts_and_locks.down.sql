-- Drop rate_lock_id from conversions
DROP INDEX IF EXISTS idx_conversions_rate_lock;
ALTER TABLE conversions DROP COLUMN IF EXISTS rate_lock_id;

-- Drop FX Rate Locks table
DROP INDEX IF EXISTS idx_fx_rate_locks_consumed;
DROP INDEX IF EXISTS idx_fx_rate_locks_expires;
DROP INDEX IF EXISTS idx_fx_rate_locks_tenant;
DROP TABLE IF EXISTS fx_rate_locks;

-- Drop FX Rate Alerts table
DROP INDEX IF EXISTS idx_fx_rate_alerts_expires;
DROP INDEX IF EXISTS idx_fx_rate_alerts_pair;
DROP INDEX IF EXISTS idx_fx_rate_alerts_tenant_active;
DROP TABLE IF EXISTS fx_rate_alerts;
