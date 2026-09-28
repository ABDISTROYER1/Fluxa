import { HttpClient } from '../http';
import {
  CreateTransferRequest,
  TransferResponse,
  ListTransactionsQuery,
  ListTransactionsResponse,
  BatchResponse,
} from '../types';

export class TransfersResource {
  constructor(private http: HttpClient) {}

  async create(
    request: CreateTransferRequest,
    options?: { signal?: AbortSignal },
  ): Promise<TransferResponse> {
    const res = await this.http.request<TransferResponse>({
      method: 'POST',
      path: '/transfers',
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async get(transferId: string, options?: { signal?: AbortSignal }): Promise<TransferResponse> {
    const res = await this.http.request<TransferResponse>({
      method: 'GET',
      path: `/transfers/${encodeURIComponent(transferId)}`,
      signal: options?.signal,
    });
    return res.data;
  }

  async list(
    query: ListTransactionsQuery,
    options?: { signal?: AbortSignal },
  ): Promise<ListTransactionsResponse> {
    const res = await this.http.request<ListTransactionsResponse>({
      method: 'GET',
      path: '/transactions',
      query,
      signal: options?.signal,
    });
    return res.data;
  }

  async createBatch(
    request: import('../types').CreateBatchRequest,
    options?: { signal?: AbortSignal },
  ): Promise<BatchResponse> {
    const res = await this.http.request<BatchResponse>({
      method: 'POST',
      path: '/transfers/batch',
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async getBatch(batchId: string, options?: { signal?: AbortSignal }): Promise<BatchResponse> {
    const res = await this.http.request<BatchResponse>({
      method: 'GET',
      path: `/transfers/batch/${encodeURIComponent(batchId)}`,
      signal: options?.signal,
    });
    return res.data;
  }

  async exportBatch(batchId: string, options?: { signal?: AbortSignal }): Promise<string> {
    const res = await this.http.request<string>({
      method: 'GET',
      path: `/transfers/batch/${encodeURIComponent(batchId)}/export`,
      signal: options?.signal,
    });
    return res.data;
  }
}
