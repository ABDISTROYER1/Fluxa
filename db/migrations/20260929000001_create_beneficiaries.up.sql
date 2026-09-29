CREATE TABLE IF NOT EXISTS beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    mode TEXT NOT NULL DEFAULT 'live',
    account TEXT NOT NULL,
    label TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'revoked')),
    cooldown_until TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT beneficiaries_tenant_mode_account_key UNIQUE (tenant_id, mode, account)
);

CREATE INDEX IF NOT EXISTS idx_beneficiaries_tenant_mode_status
    ON beneficiaries (tenant_id, mode, status);
