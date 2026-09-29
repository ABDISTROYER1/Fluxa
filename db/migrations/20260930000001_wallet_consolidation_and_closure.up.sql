-- Consolidation Operations table
CREATE TABLE IF NOT EXISTS consolidation_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_wallet_ids UUID[] NOT NULL,
    destination_wallet UUID NOT NULL REFERENCES wallets(id),
    idempotency_key VARCHAR(255) NOT NULL,
    total_fee DECIMAL(38, 18) DEFAULT 0,
    reserve_recovered DECIMAL(38, 18) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    dry_run BOOLEAN DEFAULT FALSE NOT NULL,
    transaction_hashes TEXT[],
    actor_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_consolidation_ops_tenant ON consolidation_operations(tenant_id);
CREATE INDEX idx_consolidation_ops_status ON consolidation_operations(status);
CREATE UNIQUE INDEX idx_consolidation_ops_idempotency ON consolidation_operations(tenant_id, idempotency_key);

-- Account Closures table
CREATE TABLE IF NOT EXISTS account_closures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    destination_wallet UUID NOT NULL REFERENCES wallets(id),
    actor_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    transaction_hash VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_account_closures_tenant ON account_closures(tenant_id);
CREATE INDEX idx_account_closures_wallet ON account_closures(wallet_id);
CREATE INDEX idx_account_closures_status ON account_closures(status);
