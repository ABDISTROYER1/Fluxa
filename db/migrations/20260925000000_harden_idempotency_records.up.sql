ALTER TABLE idempotency_records
    ADD COLUMN IF NOT EXISTS lease_token UUID,
    ADD COLUMN IF NOT EXISTS lease_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS response_headers JSONB,
    ADD COLUMN IF NOT EXISTS response_body_bytes BYTEA,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Rows left in processing by an older process are recoverable immediately.
-- Their original HTTP response cannot be reconstructed, so the lease is
-- intentionally backdated to the row creation time.
UPDATE idempotency_records
SET lease_expires_at = created_at,
    updated_at = NOW()
WHERE status = 'processing'
  AND lease_expires_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_idempotency_processing_lease
    ON idempotency_records (lease_expires_at)
    WHERE status = 'processing';

ALTER TABLE idempotency_records
    DROP CONSTRAINT IF EXISTS idempotency_records_status_check;
ALTER TABLE idempotency_records
    ADD CONSTRAINT idempotency_records_status_check
    CHECK (status IN ('processing', 'complete'));

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS idempotency_record_id UUID
    REFERENCES idempotency_records(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_idempotency_record
    ON transactions (idempotency_record_id)
    WHERE idempotency_record_id IS NOT NULL;
