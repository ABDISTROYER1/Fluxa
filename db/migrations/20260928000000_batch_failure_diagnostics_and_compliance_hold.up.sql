ALTER TYPE batch_status ADD VALUE IF NOT EXISTS 'compliance_hold';

ALTER TABLE transactions
    ADD COLUMN failure_reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN failure_message TEXT NOT NULL DEFAULT '';