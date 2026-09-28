ALTER TABLE transactions
    DROP COLUMN failure_reason,
    DROP COLUMN failure_message;

ALTER TABLE batches ALTER COLUMN status DROP DEFAULT;
ALTER TABLE batches ALTER COLUMN status TYPE TEXT USING status::TEXT;
DROP TYPE batch_status;
CREATE TYPE batch_status AS ENUM ('pending', 'processing', 'partial', 'completed', 'failed');
ALTER TABLE batches ALTER COLUMN status TYPE batch_status USING status::batch_status;
ALTER TABLE batches ALTER COLUMN status SET DEFAULT 'pending';