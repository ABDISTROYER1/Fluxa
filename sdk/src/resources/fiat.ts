import { HttpClient, RequestOptions, makeIdempotencyKey } from '../http';
import { DepositRequest, DepositResponse, WithdrawRequest, WithdrawResponse } from '../types';

export class FiatResource {
  constructor(private http: HttpClient) {}

  async deposit(
    walletId: string,
    request: DepositRequest,
    options?: RequestOptions,
  ): Promise<DepositResponse> {
    const res = await this.http.request<DepositResponse>({
      method: 'POST',
      path: `/wallets/${encodeURIComponent(walletId)}/deposit/fiat`,
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey ?? makeIdempotencyKey(),
    });
    return res.data;
  }

  async withdraw(
    walletId: string,
    request: WithdrawRequest,
    options?: RequestOptions,
  ): Promise<WithdrawResponse> {
    const res = await this.http.request<WithdrawResponse>({
      method: 'POST',
      path: `/wallets/${encodeURIComponent(walletId)}/withdraw/fiat`,
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey ?? makeIdempotencyKey(),
    });
    return res.data;
  }
}
