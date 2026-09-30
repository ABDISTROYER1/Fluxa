export interface Wallet {
  id: string;
  public_key: string;
  created_at: string;
}

export interface Balance {
  asset_code: string;
  issuer: string;
  balance: string;
  usd_equivalent?: string;
}

export interface WalletBalances {
  wallet_id: string;
  balances: Balance[];
}

export interface CreateWalletResponse {
  id: string;
  public_key: string;
  custody_type: string;
  contract_id?: string;
  created_at: string;
}

export interface Transaction {
  id: string;
  tx_hash?: string;
  type: string;
  status: string;
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  fee_amount: string;
  net_amount: string;
  fee_bps: number;
  failure_reason?: string;
  failure_message?: string;
  created_at: string;
}

export interface TransactionsResponse {
  transactions: Transaction[];
}

export interface CreateTransferRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
}

export interface Conversion {
  id: string;
  wallet_id: string;
  source_asset: string;
  dest_asset: string;
  source_amount: string;
  dest_amount: string;
  fee_amount: string;
  fee_bps: number;
  rate: string;
  tx_hash: string;
  created_at: string;
}

export interface FxQuoteRequest {
  from_asset: string;
  to_asset: string;
  amount: string;
}

export interface FxQuoteResponse {
  id: string;
  org_id: string;
  from_asset: string;
  to_asset: string;
  from_amount: string;
  to_amount: string;
  rate: string;
  fee: string;
  expires_at: string;
  used: boolean;
}

export interface FxConvertRequest {
  wallet_id: string;
  quote_id: string;
}

export interface FxRatesResponse {
  rate: string;
  mid_market_rate: string;
  spread_bps: number;
  provider: string;
  cached_at: string;
  stale: boolean;
}

export interface APIKey {
  id: string;
  prefix: string;
  label?: string;
  mode?: 'live' | 'test';
  scopes?: string[];
  expires_at?: string;
  rotation_reminder_days?: number;
  is_expired?: boolean;
  last_used_at?: string;
  revoked_at?: string;
  created_at: string;
}

export interface CreateAPIKeyResponse {
  id: string;
  key: string;
  prefix: string;
  label?: string;
  scopes?: string[];
  expires_at?: string;
  rotation_reminder_days?: number;
  created_at: string;
}

export interface WebhookEndpoint {
  id: string;
  url: string;
  events: string[];
  active: boolean;
  created_at: string;
  secret?: string;
}

export interface RegisterWebhookRequest {
  url: string;
  events: string[];
}

export interface WebhookDelivery {
  id: string;
  endpoint_id: string;
  event_type: string;
  payload: string;
  response_code: number;
  status: string;
  attempt_count: number;
  last_attempt: string;
  created_at: string;
}

export interface FeeSchedule {
  transfer_fee_bps: number;
  conversion_fee_bps: number;
  min_fee_amount: string;
  max_fee_amount?: string;
  asset: string;
}

export interface FiatDepositRequest {
  amount: string;
  currency: string;
  email: string;
  name: string;
}

export interface FiatDepositResponse {
  payment_link: string;
  reference: string;
}

export interface FiatWithdrawRequest {
  amount: string;
  currency: string;
  account_bank: string;
  account_number: string;
}

export interface FiatWithdrawResponse {
  reference: string;
  status: string;
}

export interface BatchTransferRequest {
  from_wallet_id: string;
  transfers: {
    to_wallet_id: string;
    asset: string;
    amount: string;
    reference?: string;
  }[];
}

export interface BatchTransferResponse {
  id: string;
  status: string;
  total_count: number;
  success_count: number;
  failed_count: number;
  held_count: number;
  created_at: string;
  transfers: {
    id: string;
    to_wallet_id: string;
    asset: string;
    amount: string;
    reference?: string;
    status: string;
    tx_hash?: string;
    failure_reason?: string;
    failure_message?: string;
  }[];
}

export type BatchResponse = BatchTransferResponse;

export interface ScheduleTransferRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: 'daily' | 'weekly' | 'monthly';
  start_date: string;
  end_date?: string;
  timezone?: string;
  missed_run_policy?: 'skip' | 'run_once';
}

export interface ScheduleTransferResponse {
  id: string;
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: string;
  timezone: string;
  missed_run_policy: 'skip' | 'run_once';
  next_run_at: string;
  end_at?: string;
  status: string;
  created_at: string;
}

export interface ScheduleRunResponse {
  id: string;
  schedule_id: string;
  expected_run_at: string;
  status: 'pending' | 'running' | 'succeeded' | 'failed' | 'skipped' | 'cancelled';
  transaction_id?: string;
  error?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface APIError {
  error: {
    code: string;
    message: string;
  };
}
