DROP INDEX IF EXISTS uq_transactions_idempotency_record;
ALTER TABLE transactions DROP COLUMN IF EXISTS idempotency_record_id;

DROP INDEX IF EXISTS idx_idempotency_processing_lease;
ALTER TABLE idempotency_records DROP CONSTRAINT IF EXISTS idempotency_records_status_check;
ALTER TABLE idempotency_records
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS response_body_bytes,
    DROP COLUMN IF EXISTS response_headers,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS lease_token;
