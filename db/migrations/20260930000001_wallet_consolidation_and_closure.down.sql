-- Drop Account Closures table
DROP INDEX IF EXISTS idx_account_closures_tenant;
DROP INDEX IF EXISTS idx_account_closures_wallet;
DROP INDEX IF EXISTS idx_account_closures_status;
DROP TABLE IF EXISTS account_closures;

-- Drop Consolidation Operations table
DROP INDEX IF EXISTS idx_consolidation_ops_idempotency;
DROP INDEX IF EXISTS idx_consolidation_ops_tenant;
DROP INDEX IF EXISTS idx_consolidation_ops_status;
DROP TABLE IF EXISTS consolidation_operations;
