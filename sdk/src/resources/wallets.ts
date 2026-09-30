import { HttpClient, RequestOptions } from '../http';
import {
  CreateTrustlineRequest,
  CreateWalletResponse,
  GetBalancesResponse,
  TrustlineResponse,
} from '../types';

export class WalletsResource {
  constructor(private http: HttpClient) {}

  async create(options?: RequestOptions): Promise<CreateWalletResponse> {
    const res = await this.http.request<CreateWalletResponse>({
      method: 'POST',
      path: '/wallets',
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }

  async getBalances(walletId: string, options?: RequestOptions): Promise<GetBalancesResponse> {
    const res = await this.http.request<GetBalancesResponse>({
      method: 'GET',
      path: `/wallets/${encodeURIComponent(walletId)}/balances`,
      signal: options?.signal,
    });
    return res.data;
  }

  async createTrustline(
    walletId: string,
    request: CreateTrustlineRequest,
    options?: RequestOptions,
  ): Promise<TrustlineResponse> {
    const res = await this.http.request<TrustlineResponse>({
      method: 'POST',
      path: `/wallets/${encodeURIComponent(walletId)}/trustlines`,
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }
}
