export type EnvironmentMode = 'live' | 'test';

export interface HealthResponse {
  status: string;
  services?: Record<string, string>;
}

export interface FeeSchedule {
  transfer_fee_bps: number;
  conversion_fee_bps: number;
  min_fee_amount: string;
  max_fee_amount?: string;
  asset: string;
}

export interface Wallet {
  id: string;
  public_key: string;
  mode: EnvironmentMode;
  custody_type?: string;
  created_at: string;
}

export interface WalletBalance {
  asset_code: string;
  issuer: string;
  balance: string;
}

export interface WalletWithBalance extends Wallet {
  balances: WalletBalance[];
}

export interface Transaction {
  id: string;
  tx_hash?: string;
  type: string;
  status: string;
  mode: EnvironmentMode;
  asset: string;
  amount: string;
  from_wallet_id: string;
  to_wallet_id: string;
  fee_amount?: string;
  net_amount?: string;
  fee_bps?: number;
  currency?: string;
  batch_id?: string;
  created_at: string;
}

export interface TransferListParams {
  before?: string;
  after?: string;
  limit?: number;
  sort?: 'created_at' | 'amount' | 'status';
  order?: 'asc' | 'desc';
  status?: string;
  date_from?: string;
  date_to?: string;
  currency?: string;
  batch_id?: string;
}

export interface CreateTransferRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
}

export interface APIKey {
  id: string;
  prefix: string;
  mode: EnvironmentMode;
  label?: string | null;
  last_used_at?: string | null;
  revoked_at?: string | null;
  created_at: string;
}

export interface WebhookEndpoint {
  id: string;
  url: string;
  events: string[];
  mode: EnvironmentMode;
  active: boolean;
}

export interface WebhookDelivery {
  id: string;
  endpoint_id: string;
  event_type?: string;
  status: string;
  status_code?: number;
  response_code?: number;
  attempt_count?: number;
  created_at: string;
}

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: string;
  status: string;
  created_at: string;
  resolved_at?: string;
}

export interface StatusResponse {
  api_version: string;
  status: string;
  message: string;
  recent_incidents: Incident[];
}

export interface FeeCollectedSummary {
  collected?: Array<{
    asset: string;
    total_fees: string;
    transfer_count: number;
  }>;
}

// --- FX ---
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
}

// --- Batch ---
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

export interface BatchResponse {
  id: string;
  status: string;
  total_count: number;
  success_count: number;
  failed_count: number;
  created_at: string;
  transfers?: Array<{
    id: string;
    to_wallet_id: string;
    asset: string;
    amount: string;
    reference?: string;
    status: string;
    tx_hash?: string;
  }>;
}

// --- Schedules ---
export interface CreateScheduleRequest {
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: 'daily' | 'weekly' | 'monthly';
  start_date: string;
  end_date?: string;
}

export interface UpdateScheduleRequest {
  status?: 'active' | 'paused';
  amount?: string;
  frequency?: 'daily' | 'weekly' | 'monthly';
  end_date?: string;
}

export interface ScheduleResponse {
  id: string;
  from_wallet_id: string;
  to_wallet_id: string;
  asset: string;
  amount: string;
  frequency: string;
  next_run_at: string;
  end_at?: string;
  status: string;
  created_at: string;
}

// --- Fiat ---
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

// --- Trustline ---
export interface CreateTrustlineRequest {
  asset: string;
  issuer?: string;
  limit?: string;
}

export interface TrustlineResponse {
  wallet_id: string;
  asset: string;
  asset_code?: string;
  asset_issuer?: string;
  issuer?: string;
  status: string;
  tx_hash?: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';

function currentMode(): EnvironmentMode {
  if (typeof window === 'undefined') return 'live';
  return window.localStorage.getItem('fluxa_mode') === 'test' ? 'test' : 'live';
}

function storedToken(): string | null {
  if (typeof window === 'undefined') return null;
  return window.localStorage.getItem('fluxa_api_key') || window.localStorage.getItem('fluxa_token');
}

async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const token = storedToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string> | undefined),
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  // API-key authentication derives mode from the key itself. For JWT-based
  // dashboard sessions this header selects the explicitly displayed profile.
  headers['X-Fluxa-Mode'] = currentMode();

  const res = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });
  if (res.status === 204) return {} as T;

  const text = await res.text();
  let body: unknown = undefined;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = undefined;
    }
  }
  if (!res.ok) {
    const parsed = body as { error?: { message?: string }; message?: string } | undefined;
    const message = parsed?.error?.message || parsed?.message || text || `API error: ${res.status}`;
    throw new Error(message);
  }
  return (body ?? {}) as T;
}

export const api = {
  getHealth: () => request<HealthResponse>('/health'),
  getFeeSchedule: () => request<FeeSchedule>('/v1/fees/'),
  listFeeCollected: (startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.set('start_date', startDate);
    if (endDate) params.set('end_date', endDate);
    const qs = params.toString();
    return request<FeeCollectedSummary>(`/v1/admin/fees/collected${qs ? `?${qs}` : ''}`);
  },
  listWallets: () => request<{ wallets: Wallet[] }>('/v1/wallets/'),
  createWallet: () => request<Wallet>('/v1/wallets/', { method: 'POST' }),
  getWallet: (id: string) => request<Wallet>(`/v1/wallets/${id}`),
  getWalletBalances: (id: string) => request<{ wallet_id?: string; balances: WalletBalance[] }>(`/v1/wallets/${id}/balances`),
  createTrustline: (walletId: string, data: CreateTrustlineRequest) => request<TrustlineResponse>(`/v1/wallets/${walletId}/trustlines`, {
    method: 'POST', body: JSON.stringify(data),
  }),
  listTransactions: (walletId: string, params: number | TransferListParams = 10) => {
    const query = typeof params === 'number' ? { limit: params } : params;
    const search = new URLSearchParams(
      Object.entries(query)
        .filter(([, value]) => value !== undefined)
        .map(([key, value]) => [key, String(value)]),
    );
    search.set('wallet_id', walletId);
    return request<{ transactions: Transaction[]; next_cursor?: string; has_more?: boolean }>(`/v1/transactions?${search}`);
  },
  createTransfer: (data: CreateTransferRequest) => request<Transaction>('/v1/transfers/', {
    method: 'POST', body: JSON.stringify(data),
  }),
  getTransaction: (id: string) => request<Transaction>(`/v1/transfers/${id}`),

  // FX
  getQuote: (data: QuoteRequest) => request<QuoteResponse>('/v1/fx/quote', {
    method: 'POST', body: JSON.stringify(data),
  }),
  convert: (data: ConvertRequest) => request<ConversionResponse>('/v1/fx/convert', {
    method: 'POST', body: JSON.stringify(data),
  }),
  getRates: (from: string, to: string) => request<RateResponse>(`/v1/fx/rates?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),

  // Batch
  createBatch: (data: CreateBatchRequest) => request<BatchResponse>('/v1/transfers/batch/', {
    method: 'POST', body: JSON.stringify(data),
  }),
  getBatch: (batchId: string) => request<BatchResponse>(`/v1/transfers/batch/${batchId}`),
  exportBatchCsv: async (batchId: string): Promise<string> => {
    const token = storedToken();
    const headers: Record<string, string> = { 'X-Fluxa-Mode': currentMode() };
    if (token) headers.Authorization = `Bearer ${token}`;
    const res = await fetch(`${API_BASE}/v1/transfers/batch/${batchId}/export`, { headers });
    if (!res.ok) throw new Error(`Export failed (${res.status})`);
    return res.text();
  },

  // Schedules
  createSchedule: (data: CreateScheduleRequest) => request<ScheduleResponse>('/v1/schedules/', {
    method: 'POST', body: JSON.stringify(data),
  }),
  listSchedules: () => request<{ schedules: ScheduleResponse[] }>('/v1/schedules/'),
  updateSchedule: (id: string, data: UpdateScheduleRequest) => request<ScheduleResponse>(`/v1/schedules/${id}`, {
    method: 'PATCH', body: JSON.stringify(data),
  }),
  cancelSchedule: (id: string) => request<void>(`/v1/schedules/${id}`, { method: 'DELETE' }),

  // Fiat
  fiatDeposit: (walletId: string, data: FiatDepositRequest) => request<FiatDepositResponse>(`/v1/wallets/${walletId}/deposit/fiat`, {
    method: 'POST', body: JSON.stringify(data),
  }),
  fiatWithdraw: (walletId: string, data: FiatWithdrawRequest) => request<FiatWithdrawResponse>(`/v1/wallets/${walletId}/withdraw/fiat`, {
    method: 'POST', body: JSON.stringify(data),
  }),

  listWebhooks: () => request<{ endpoints: WebhookEndpoint[] }>('/v1/webhooks/'),
  registerWebhook: (url: string, events: string[]) => request<WebhookEndpoint>('/v1/webhooks/', {
    method: 'POST', body: JSON.stringify({ url, events }),
  }),
  deleteWebhook: (id: string) => request<void>(`/v1/webhooks/${id}`, { method: 'DELETE' }),
  listDeliveries: (endpointId: string, limit = 10) => request<{ deliveries: WebhookDelivery[] }>(`/v1/webhooks/${endpointId}/deliveries?limit=${limit}`),
  getWebhookSecret: () => request<{ signing_secret: string }>('/v1/webhooks/secret'),
  rotateWebhookSecret: () => request<{ signing_secret: string }>('/v1/webhooks/secret/rotate', { method: 'POST' }),
  verifyWebhookSignature: (payload: { secret: string; timestamp: string; body: string; signature: string }) => request<{ valid: boolean; reason: string | null }>('/v1/webhooks/verify', {
    method: 'POST', body: JSON.stringify(payload),
  }),

  listAPIKeys: () => request<APIKey[]>('/v1/keys/'),
  createAPIKey: (label?: string, mode: EnvironmentMode = currentMode()) => request<APIKey & { key: string }>('/v1/keys/', {
    method: 'POST', body: JSON.stringify({ label, mode }),
  }),
  revokeAPIKey: (id: string) => request<void>(`/v1/keys/${id}`, { method: 'DELETE' }),

  // Test mode only: fire a simulated event to registered webhooks.
  triggerTestEvent: (event: 'transfer.settled' | 'transfer.failed' | 'wallet.funded', data: Record<string, unknown> = {}) => request<{ event: string; status: string }>('/v1/test/trigger-event', {
    method: 'POST', body: JSON.stringify({ event, data }),
  }),

  // Public platform status (no tenant authentication).
  getStatus: () => request<StatusResponse>('/status'),
  listIncidents: () => request<{ incidents: Incident[] }>('/status/incidents'),
};
