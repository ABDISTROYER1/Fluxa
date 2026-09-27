-- Add an index on expires_at so the background cleanup job can delete expired
-- idempotency_records without a sequential scan. The partial predicate limits
-- the index to rows that are expired or about to expire, keeping it small.
--
-- Mirrors idx_revoked_refresh_tokens_expires_at added in 000024.
CREATE INDEX idx_idempotency_records_expires_at
    ON idempotency_records (expires_at);
