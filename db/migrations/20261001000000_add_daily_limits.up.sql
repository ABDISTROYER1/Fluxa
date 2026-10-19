ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_transfers_per_day INT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_withdrawals_per_day INT;
