DROP INDEX IF EXISTS uq_idempotency_org_mode_key;
DROP INDEX IF EXISTS uq_conversions_mode_tx_hash;
DROP INDEX IF EXISTS uq_transactions_mode_tx_hash;
DROP INDEX IF EXISTS uq_wallets_mode_public_key;
DROP INDEX IF EXISTS idx_fx_quotes_org_mode;
DROP INDEX IF EXISTS idx_schedules_tenant_mode;
DROP INDEX IF EXISTS idx_batches_tenant_mode;
DROP INDEX IF EXISTS idx_webhook_endpoints_tenant_mode;
DROP INDEX IF EXISTS idx_transactions_tenant_mode;
DROP INDEX IF EXISTS idx_wallets_tenant_mode;
DROP INDEX IF EXISTS idx_api_keys_tenant_mode;

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
        EXECUTE format('ALTER TABLE %I DROP COLUMN IF EXISTS mode', table_name);
    END LOOP;
END $$;

UPDATE api_keys SET prefix = LEFT(prefix, 8);
ALTER TABLE api_keys ALTER COLUMN prefix TYPE VARCHAR(8);

ALTER TABLE wallets ADD CONSTRAINT wallets_public_key_key UNIQUE (public_key);
ALTER TABLE transactions ADD CONSTRAINT transactions_tx_hash_key UNIQUE (tx_hash);
ALTER TABLE conversions ADD CONSTRAINT conversions_tx_hash_key UNIQUE (tx_hash);
CREATE UNIQUE INDEX idx_idempotency_records_org_key ON idempotency_records (org_id, key);
