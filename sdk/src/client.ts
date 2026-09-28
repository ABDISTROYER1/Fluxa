import { HttpClient, HttpClientConfig } from './http';
import { WalletsResource } from './resources/wallets';
import { TransfersResource } from './resources/transfers';
import { FXResource } from './resources/fx';
import { SchedulesResource } from './resources/schedules';
import { WebhooksResource } from './resources/webhooks';
import { FeesResource } from './resources/fees';
import { KeysResource } from './resources/keys';
import { FiatResource } from './resources/fiat';

export interface FluxaClientConfig {
  apiKey: string;
  baseUrl?: string;
  timeout?: number;
  maxRetries?: number;
  retryDelay?: number;
}

const DEFAULT_BASE_URL = 'https://api.fluxa.io';
const DEFAULT_TIMEOUT = 30_000;
const DEFAULT_MAX_RETRIES = 0;
const DEFAULT_RETRY_DELAY = 500;

export class FluxaClient {
  readonly wallets: WalletsResource;
  readonly transfers: TransfersResource;
  readonly fx: FXResource;
  readonly schedules: SchedulesResource;
  readonly webhooks: WebhooksResource;
  readonly fees: FeesResource;
  readonly keys: KeysResource;
  readonly fiat: FiatResource;

  private http: HttpClient;

  constructor(config: FluxaClientConfig) {
    if (!config.apiKey) {
      throw new Error('apiKey is required');
    }

    const httpConfig: HttpClientConfig = {
      baseUrl: config.baseUrl ?? DEFAULT_BASE_URL,
      apiKey: config.apiKey,
      timeout: config.timeout && config.timeout > 0 ? config.timeout : DEFAULT_TIMEOUT,
      maxRetries: Math.max(0, config.maxRetries ?? DEFAULT_MAX_RETRIES),
      retryDelay: Math.max(0, config.retryDelay ?? DEFAULT_RETRY_DELAY),
    };

    this.http = new HttpClient(httpConfig);
    this.wallets = new WalletsResource(this.http);
    this.transfers = new TransfersResource(this.http);
    this.fx = new FXResource(this.http);
    this.schedules = new SchedulesResource(this.http);
    this.webhooks = new WebhooksResource(this.http);
    this.fees = new FeesResource(this.http);
    this.keys = new KeysResource(this.http);
    this.fiat = new FiatResource(this.http);
  }

  async health(options?: { signal?: AbortSignal }): Promise<{
    status: string;
    services?: Record<string, string>;
  }> {
    const res = await this.http.request<{
      status: string;
      services?: Record<string, string>;
    }>({
      method: 'GET',
      path: '/../health',
      signal: options?.signal,
    });
    return res.data;
  }
}
