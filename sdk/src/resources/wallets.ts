import { HttpClient } from '../http';
import { CreateWalletResponse, GetBalancesResponse } from '../types';

export class WalletsResource {
  constructor(private http: HttpClient) {}

  async create(options?: { signal?: AbortSignal }): Promise<CreateWalletResponse> {
    const res = await this.http.request<CreateWalletResponse>({
      method: 'POST',
      path: '/wallets',
      signal: options?.signal,
    });
    return res.data;
  }

  async getBalances(
    walletId: string,
    options?: { signal?: AbortSignal },
  ): Promise<GetBalancesResponse> {
    const res = await this.http.request<GetBalancesResponse>({
      method: 'GET',
      path: `/wallets/${encodeURIComponent(walletId)}/balances`,
      signal: options?.signal,
    });
    return res.data;
  }
}
