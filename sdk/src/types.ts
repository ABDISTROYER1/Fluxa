// ── Wallet ──────────────────────────────────────────────────────────────────

export interface CreateWalletResponse {
  id: string;
  public_key: string;
  created_at: string;
}

export interface Balance {
  asset_code: string;
  issuer: string;
  balance: string;
}

export interface GetBalancesResponse {
  wallet_id: string;
  balances: Balance[];
}

// ── Transaction / Transfer ──────────────────────────────────────────────────

export type TransactionStatus =
  | "pending"
  | "submitted"
  | "confirmed"
  | "failed"
  | "reconciliation_failed";

export type TransactionType = "transfer" | "conversion" | "funding";

export interface CreateTransferRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
}

export interface TransferResponse {
  id: string;
  tx_hash?: string;
  type: TransactionType;
  status: TransactionStatus;
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

export interface ListTransactionsQuery {
  wallet_id: string;
  limit?: number;
  offset?: number;
}

export interface ListTransactionsResponse {
  transactions: TransferResponse[];
}

// ── Batch ───────────────────────────────────────────────────────────────────

export type BatchStatus =
  | "pending"
  | "processing"
  | "partial"
  | "completed"
  | "failed"
  | "compliance_hold";

export interface BatchItemRequest {
  to_wallet_id: string;
  asset: string;
  amount: string;
  reference?: string;
}

export interface CreateBatchRequest {
  from_wallet_id: string;
  transfers: BatchItemRequest[];
}

export interface BatchTransferResponse {
  id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  reference?: string;
  status: TransactionStatus;
  tx_hash?: string;
  failure_reason?: string;
  failure_message?: string;
}

export interface BatchResponse {
  id: string;
  status: BatchStatus;
  total_count: number;
  success_count: number;
  failed_count: number;
  held_count: number;
  created_at: string;
  transfers?: BatchTransferResponse[];
}

// ── FX ──────────────────────────────────────────────────────────────────────

export interface QuoteRequest {
  from_asset: string;
  to_asset: string;
  amount: string;
}

export interface QuoteResponse {
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

export interface ConvertRequest {
  wallet_id: string;
  quote_id: string;
}

export interface ConversionResponse {
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

export interface RateResponse {
  rate: string;
  mid_market_rate: string;
  spread_bps: number;
  provider: string;
  cached_at: string;
  stale: boolean;
  source_amount: string;
  dest_amount: string;
  fee_amount: string;
  net_amount: string;
  fee_bps: number;
}

export interface GetRatesQuery {
  from: string;
  to: string;
}

// ── Fees ────────────────────────────────────────────────────────────────────

export interface FeeScheduleResponse {
  transfer_fee_bps: number;
  conversion_fee_bps: number;
  min_fee_amount: string;
  max_fee_amount?: string;
  asset: string;
}

export interface FeeCollectionSummary {
  asset: string;
  total_fees: string;
  tenant_fees: TenantFeeTotal[];
}

export interface TenantFeeTotal {
  tenant_id?: string;
  total_fees: string;
}

export interface ListCollectedQuery {
  start_date?: string;
  end_date?: string;
}

export interface ListCollectedResponse {
  summary: FeeCollectionSummary[];
}

// ── Fiat ────────────────────────────────────────────────────────────────────

export interface DepositRequest {
  amount: string;
  currency: string;
  email: string;
  name: string;
}

export interface DepositResponse {
  payment_link: string;
  reference: string;
}

export interface WithdrawRequest {
  amount: string;
  currency: string;
  account_bank: string;
  account_number: string;
}

export interface WithdrawResponse {
  reference: string;
  status: string;
}

// ── Webhook ─────────────────────────────────────────────────────────────────

export type EventType =
  | "transfer.initiated"
  | "transfer.settled"
  | "transfer.failed"
  | "wallet.funded"
  | "conversion.completed";

export type DeliveryStatus = "pending" | "success" | "failed";

export interface RegisterWebhookRequest {
  url: string;
  events?: string[];
}

export interface WebhookEndpointResponse {
  id: string;
  url: string;
  secret?: string;
  events: string[];
  active: boolean;
  created_at: string;
}

export interface ListWebhooksResponse {
  endpoints: WebhookEndpointResponse[];
}

export interface WebhookDeliveryResponse {
  id: string;
  endpoint_id: string;
  event_type: string;
  status: DeliveryStatus;
  response_code?: number;
  attempt_count: number;
  last_attempt?: string;
  created_at: string;
}

export interface ListDeliveriesResponse {
  deliveries: WebhookDeliveryResponse[];
}

// ── Schedule ────────────────────────────────────────────────────────────────

export type ScheduleFrequency = "daily" | "weekly" | "monthly";

export type ScheduleStatus = "active" | "paused" | "cancelled" | "completed";

export interface CreateScheduleRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: ScheduleFrequency;
  start_date: string;
  end_date?: string;
}

export interface UpdateScheduleRequest {
  status?: "active" | "paused";
  amount?: string;
  frequency?: ScheduleFrequency;
  end_date?: string;
}

export interface ScheduleResponse {
  id: string;
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: ScheduleFrequency;
  next_run_at: string;
  end_at?: string;
  status: ScheduleStatus;
  created_at: string;
}

export interface ListSchedulesResponse {
  schedules: ScheduleResponse[];
}

// ── API Key ─────────────────────────────────────────────────────────────────

export interface CreateKeyRequest {
  label?: string;
}

export interface CreateKeyResponse {
  id: string;
  key: string;
  prefix: string;
  label?: string;
  created_at: string;
}

export interface APIKeyResponse {
  id: string;
  prefix: string;
  label?: string;
  last_used_at?: string;
  revoked_at?: string;
  created_at: string;
}

// ── Health ──────────────────────────────────────────────────────────────────

export interface HealthResponse {
  status: string;
  services?: Record<string, string>;
}
