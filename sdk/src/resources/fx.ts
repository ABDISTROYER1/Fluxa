import { HttpClient } from '../http';
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

  async quote(request: QuoteRequest, options?: { signal?: AbortSignal }): Promise<QuoteResponse> {
    const res = await this.http.request<QuoteResponse>({
      method: 'POST',
      path: '/fx/quote',
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async convert(
    request: ConvertRequest,
    options?: { signal?: AbortSignal },
  ): Promise<ConversionResponse> {
    const res = await this.http.request<ConversionResponse>({
      method: 'POST',
      path: '/fx/convert',
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async getRates(query: GetRatesQuery, options?: { signal?: AbortSignal }): Promise<RateResponse> {
    const res = await this.http.request<RateResponse>({
      method: 'GET',
      path: '/fx/rates',
      query,
      signal: options?.signal,
    });
    return res.data;
  }
}
