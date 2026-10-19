import { HttpClient, RequestOptions } from '../http';
import { randomUUID } from 'crypto';
import {
  QuoteRequest,
  QuoteResponse,
  ConvertRequest,
  ConversionResponse,
  GetRatesQuery,
  RateResponse,
} from '../types';

export class FXResource {
  constructor(private http: HttpClient) {}

  async quote(request: QuoteRequest, options?: RequestOptions): Promise<QuoteResponse> {
    const idempotencyKey = options?.idempotencyKey ?? randomUUID();
    const res = await this.http.request<QuoteResponse>({
      method: 'POST',
      path: '/fx/quote',
      body: request,
      signal: options?.signal,
      idempotencyKey,
    });
    return res.data;
  }

  async convert(request: ConvertRequest, options?: RequestOptions): Promise<ConversionResponse> {
    const idempotencyKey = options?.idempotencyKey ?? randomUUID();
    const res = await this.http.request<ConversionResponse>({
      method: 'POST',
      path: '/fx/convert',
      body: request,
      signal: options?.signal,
      idempotencyKey,
    });
    return res.data;
  }

  async getRates(query: GetRatesQuery, options?: RequestOptions): Promise<RateResponse> {
    const res = await this.http.request<RateResponse>({
      method: 'GET',
      path: '/fx/rates',
      query,
      signal: options?.signal,
    });
    return res.data;
  }
}
