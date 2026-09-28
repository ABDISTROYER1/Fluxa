import { HttpClient, RequestOptions } from '../http';
import { CreateKeyRequest, CreateKeyResponse, APIKeyResponse } from '../types';

export class KeysResource {
  constructor(private http: HttpClient) {}

  async create(request?: CreateKeyRequest, options?: RequestOptions): Promise<CreateKeyResponse> {
    const res = await this.http.request<CreateKeyResponse>({
      method: 'POST',
      path: '/keys',
      body: request ?? {},
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }

  async list(options?: RequestOptions): Promise<APIKeyResponse[]> {
    const res = await this.http.request<APIKeyResponse[]>({
      method: 'GET',
      path: '/keys',
      signal: options?.signal,
    });
    return res.data;
  }

  async delete(keyId: string, options?: RequestOptions): Promise<void> {
    await this.http.request<unknown>({
      method: 'DELETE',
      path: `/keys/${encodeURIComponent(keyId)}`,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
  }
}
