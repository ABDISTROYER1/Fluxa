import { HttpClient, RequestOptions } from '../http';
import { FeeScheduleResponse, ListCollectedQuery, ListCollectedResponse } from '../types';

export class FeesResource {
  constructor(private http: HttpClient) {}

  async get(options?: RequestOptions): Promise<FeeScheduleResponse> {
    const res = await this.http.request<FeeScheduleResponse>({
      method: 'GET',
      path: '/fees',
      signal: options?.signal,
    });
    return res.data;
  }

  async listCollected(
    query?: ListCollectedQuery,
    options?: RequestOptions,
  ): Promise<ListCollectedResponse> {
    const res = await this.http.request<ListCollectedResponse>({
      method: 'GET',
      path: '/admin/fees/collected',
      query,
      signal: options?.signal,
    });
    return res.data;
  }
}
