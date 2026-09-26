CREATE INDEX idx_transactions_reconciliation ON transactions (status, created_at, reconciled_at);
