-- Existing rows are live by definition. New request-path resources copy the
-- authenticated key mode into these columns before persistence.
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE idempotency_records ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE webhook_endpoints ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE webhook_deliveries ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE webhook_subscriptions ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE webhook_dead_letters ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE batches ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE schedules ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE conversions ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE fx_quotes ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE fees ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';
ALTER TABLE fee_collections ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'live';

ALTER TABLE api_keys ALTER COLUMN prefix TYPE VARCHAR(16);

DO $$
DECLARE
    table_name TEXT;
BEGIN
    FOREACH table_name IN ARRAY ARRAY[
        'api_keys', 'wallets', 'transactions', 'idempotency_records',
        'webhook_endpoints', 'webhook_deliveries', 'webhook_subscriptions',
        'webhook_dead_letters',
        'batches', 'schedules', 'conversions', 'fx_quotes', 'fees',
        'fee_collections'
    ]
    LOOP
        EXECUTE format('ALTER TABLE %I DROP CONSTRAINT IF EXISTS %I', table_name, table_name || '_mode_check');
        EXECUTE format(
            'ALTER TABLE %I ADD CONSTRAINT %I CHECK (mode IN (''live'', ''test''))',
            table_name,
            table_name || '_mode_check'
        );
    END LOOP;
END $$;

CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_mode ON api_keys (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_wallets_tenant_mode ON wallets (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_transactions_tenant_mode ON transactions (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_tenant_mode ON webhook_endpoints (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_batches_tenant_mode ON batches (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_schedules_tenant_mode ON schedules (tenant_id, mode);
CREATE INDEX IF NOT EXISTS idx_fx_quotes_org_mode ON fx_quotes (org_id, mode);

-- Stellar identifiers are network-specific. The same hash/account may exist
-- independently in public and testnet.
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_public_key_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_wallets_mode_public_key
    ON wallets (mode, public_key);

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_tx_hash_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_mode_tx_hash
    ON transactions (mode, tx_hash)
    WHERE tx_hash IS NOT NULL;

ALTER TABLE conversions DROP CONSTRAINT IF EXISTS conversions_tx_hash_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_conversions_mode_tx_hash
    ON conversions (mode, tx_hash)
    WHERE tx_hash IS NOT NULL;

DROP INDEX IF EXISTS idx_idempotency_records_org_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_idempotency_org_mode_key
    ON idempotency_records (org_id, mode, key);
