import { HttpClient, RequestOptions } from '../http';
import { CreatePaymentLinkRequest, PaymentLinkResponse, PaymentLinksResponse } from '../types';

export class PaymentLinksResource {
  constructor(private http: HttpClient) {}

  async create(
    request: CreatePaymentLinkRequest,
    options?: RequestOptions,
  ): Promise<PaymentLinkResponse> {
    const response = await this.http.request<PaymentLinkResponse>({
      method: 'POST',
      path: '/payment-links',
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return response.data;
  }

  async list(options?: RequestOptions): Promise<PaymentLinksResponse> {
    const response = await this.http.request<PaymentLinksResponse>({
      method: 'GET',
      path: '/payment-links',
      signal: options?.signal,
    });
    return response.data;
  }

  async get(id: string, options?: RequestOptions): Promise<PaymentLinkResponse> {
    const response = await this.http.request<PaymentLinkResponse>({
      method: 'GET',
      path: `/payment-links/${encodeURIComponent(id)}`,
      signal: options?.signal,
    });
    return response.data;
  }

  async cancel(id: string, options?: RequestOptions): Promise<void> {
    await this.http.request({
      method: 'DELETE',
      path: `/payment-links/${encodeURIComponent(id)}`,
      signal: options?.signal,
    });
  }
}
