import { HttpClient, RequestOptions } from '../http';
import { CreateRefundRequest, RefundResponse, RefundsResponse } from '../types';

export class RefundsResource {
  constructor(private http: HttpClient) {}

  async create(request: CreateRefundRequest, options?: RequestOptions): Promise<RefundResponse> {
    const response = await this.http.request<RefundResponse>({
      method: 'POST',
      path: '/refunds',
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return response.data;
  }

  async get(id: string, options?: RequestOptions): Promise<RefundResponse> {
    const response = await this.http.request<RefundResponse>({
      method: 'GET',
      path: `/refunds/${encodeURIComponent(id)}`,
      signal: options?.signal,
    });
    return response.data;
  }

  async list(originalTransactionId: string, options?: RequestOptions): Promise<RefundsResponse> {
    const response = await this.http.request<RefundsResponse>({
      method: 'GET',
      path: '/refunds',
      query: { original_transaction_id: originalTransactionId },
      signal: options?.signal,
    });
    return response.data;
  }
}
