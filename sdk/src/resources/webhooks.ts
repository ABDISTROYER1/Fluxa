import { HttpClient, RequestOptions } from '../http';
import {
  RegisterWebhookRequest,
  WebhookEndpointResponse,
  ListWebhooksResponse,
  ListDeliveriesResponse,
  ListWebhookSigningSecretsResponse,
  RotateWebhookSigningSecretRequest,
  RotateWebhookSigningSecretResponse,
} from '../types';

export class WebhooksResource {
  constructor(private http: HttpClient) {}

  async create(
    request: RegisterWebhookRequest,
    options?: RequestOptions,
  ): Promise<WebhookEndpointResponse> {
    const res = await this.http.request<WebhookEndpointResponse>({
      method: 'POST',
      path: '/webhooks',
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }

  async list(options?: RequestOptions): Promise<ListWebhooksResponse> {
    const res = await this.http.request<ListWebhooksResponse>({
      method: 'GET',
      path: '/webhooks',
      signal: options?.signal,
    });
    return res.data;
  }

  async delete(webhookId: string, options?: RequestOptions): Promise<void> {
    await this.http.request<unknown>({
      method: 'DELETE',
      path: `/webhooks/${encodeURIComponent(webhookId)}`,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
  }

  async getDeliveries(
    webhookId: string,
    query?: { limit?: number; offset?: number },
    options?: RequestOptions,
  ): Promise<ListDeliveriesResponse> {
    const res = await this.http.request<ListDeliveriesResponse>({
      method: 'GET',
      path: `/webhooks/${encodeURIComponent(webhookId)}/deliveries`,
      query,
      signal: options?.signal,
    });
    return res.data;
  }

  async getSigningSecrets(options?: RequestOptions): Promise<ListWebhookSigningSecretsResponse> {
    const res = await this.http.request<ListWebhookSigningSecretsResponse>({
      method: 'GET',
      path: '/webhooks/secret',
      signal: options?.signal,
    });
    return res.data;
  }

  async rotateSigningSecret(
    request: RotateWebhookSigningSecretRequest = {},
    options?: RequestOptions,
  ): Promise<RotateWebhookSigningSecretResponse> {
    const res = await this.http.request<RotateWebhookSigningSecretResponse>({
      method: 'POST',
      path: '/webhooks/secret/rotate',
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }
}
