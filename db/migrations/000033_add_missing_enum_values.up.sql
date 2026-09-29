ALTER TYPE schedule_status ADD VALUE IF NOT EXISTS 'processing';
ALTER TYPE schedule_status ADD VALUE IF NOT EXISTS 'failed';
ALTER TYPE batch_status ADD VALUE IF NOT EXISTS 'compliance_hold';
