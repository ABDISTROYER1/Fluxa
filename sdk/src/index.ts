export { FluxaClient } from './client';
export type { FluxaClientConfig } from './client';

export {
  FluxaError,
  AuthenticationError,
  NotFoundError,
  ValidationError,
  RateLimitError,
  ConflictError,
} from './errors';

export type {
  // Wallet
  CreateWalletResponse,
  Balance,
  GetBalancesResponse,
  // Transfer
  TransactionStatus,
  TransactionType,
  CreateTransferRequest,
  TransferResponse,
  ListTransactionsQuery,
  ListTransactionsResponse,
  // Batch
  BatchStatus,
  BatchItemRequest,
  CreateBatchRequest,
  BatchTransferResponse,
  BatchResponse,
  // FX
  QuoteRequest,
  QuoteResponse,
  ConvertRequest,
  ConversionResponse,
  GetRatesQuery,
  RateResponse,
  // Fees
  FeeScheduleResponse,
  FeeCollectionSummary,
  TenantFeeTotal,
  ListCollectedQuery,
  ListCollectedResponse,
  // Fiat
  DepositRequest,
  DepositResponse,
  WithdrawRequest,
  WithdrawResponse,
  // Webhook
  EventType,
  DeliveryStatus,
  RegisterWebhookRequest,
  WebhookEndpointResponse,
  ListWebhooksResponse,
  WebhookDeliveryResponse,
  ListDeliveriesResponse,
  // Schedule
  ScheduleFrequency,
  ScheduleStatus,
  CreateScheduleRequest,
  UpdateScheduleRequest,
  ScheduleResponse,
  ListSchedulesResponse,
  ScheduleRunStatus,
  ScheduleRunResponse,
  ListScheduleRunsResponse,
  // API Key
  CreateKeyRequest,
  CreateKeyResponse,
  APIKeyResponse,
  // Health
  HealthResponse,
} from './types';

// Re-export resource classes for advanced usage
export { WalletsResource } from './resources/wallets';
export { TransfersResource } from './resources/transfers';
export { FXResource } from './resources/fx';
export { SchedulesResource } from './resources/schedules';
export { WebhooksResource } from './resources/webhooks';
export { FeesResource } from './resources/fees';
export { KeysResource } from './resources/keys';
export { FiatResource } from './resources/fiat';
