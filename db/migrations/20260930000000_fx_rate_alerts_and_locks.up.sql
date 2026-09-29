-- FX Rate Alerts table
CREATE TABLE IF NOT EXISTS fx_rate_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    from_asset VARCHAR(12) NOT NULL,
    to_asset VARCHAR(12) NOT NULL,
    target_rate DECIMAL(38, 18) NOT NULL,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('at_or_above', 'at_or_below')),
    expires_at TIMESTAMP,
    last_fired_at TIMESTAMP,
    last_eval_rate DECIMAL(38, 18),
    active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fx_rate_alerts_tenant_active ON fx_rate_alerts(tenant_id, active) WHERE active = TRUE;
CREATE INDEX idx_fx_rate_alerts_pair ON fx_rate_alerts(from_asset, to_asset, active) WHERE active = TRUE;
CREATE INDEX idx_fx_rate_alerts_expires ON fx_rate_alerts(expires_at) WHERE expires_at IS NOT NULL AND active = TRUE;

-- FX Rate Locks table
CREATE TABLE IF NOT EXISTS fx_rate_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    from_asset VARCHAR(12) NOT NULL,
    to_asset VARCHAR(12) NOT NULL,
    locked_rate DECIMAL(38, 18) NOT NULL,
    amount DECIMAL(38, 18) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    consumed_at TIMESTAMP,
    consumed_by UUID REFERENCES conversions(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fx_rate_locks_tenant ON fx_rate_locks(tenant_id);
CREATE INDEX idx_fx_rate_locks_expires ON fx_rate_locks(expires_at) WHERE consumed_at IS NULL;
CREATE INDEX idx_fx_rate_locks_consumed ON fx_rate_locks(consumed_at);

-- Add rate_lock_id column to conversions table
ALTER TABLE conversions ADD COLUMN IF NOT EXISTS rate_lock_id UUID REFERENCES fx_rate_locks(id);
CREATE INDEX IF NOT EXISTS idx_conversions_rate_lock ON conversions(rate_lock_id) WHERE rate_lock_id IS NOT NULL;
