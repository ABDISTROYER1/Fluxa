import { HttpClient } from '../http';
import { CreateKeyRequest, CreateKeyResponse, APIKeyResponse } from '../types';

export class KeysResource {
  constructor(private http: HttpClient) {}

  async create(
    request?: CreateKeyRequest,
    options?: { signal?: AbortSignal },
  ): Promise<CreateKeyResponse> {
    const res = await this.http.request<CreateKeyResponse>({
      method: 'POST',
      path: '/keys',
      body: request ?? {},
      signal: options?.signal,
    });
    return res.data;
  }

  async list(options?: { signal?: AbortSignal }): Promise<APIKeyResponse[]> {
    const res = await this.http.request<APIKeyResponse[]>({
      method: 'GET',
      path: '/keys',
      signal: options?.signal,
    });
    return res.data;
  }

  async delete(keyId: string, options?: { signal?: AbortSignal }): Promise<void> {
    await this.http.request<unknown>({
      method: 'DELETE',
      path: `/keys/${encodeURIComponent(keyId)}`,
      signal: options?.signal,
    });
  }
}
