CREATE TABLE fee_tiers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID,
    min_volume         NUMERIC(20, 7) NOT NULL,
    transfer_fee_bps   INT NOT NULL,
    conversion_fee_bps INT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fee_tiers_tenant ON fee_tiers (tenant_id, min_volume DESC);
