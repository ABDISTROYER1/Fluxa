import { HttpClient, RequestOptions } from '../http';
import {
  CreateScheduleRequest,
  UpdateScheduleRequest,
  ScheduleResponse,
  ListSchedulesResponse,
  ListScheduleRunsResponse,
} from '../types';

function newIdempotencyKey(): string {
  if (globalThis.crypto?.randomUUID) return globalThis.crypto.randomUUID();

  const bytes = new Uint8Array(16);
  if (globalThis.crypto?.getRandomValues) {
    globalThis.crypto.getRandomValues(bytes);
  } else {
    // Idempotency keys are collision-avoidance tokens, not credentials.
    for (let i = 0; i < bytes.length; i++) {
      bytes[i] = Math.floor(Math.random() * 256);
    }
  }
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0'));
  return [
    hex.slice(0, 4).join(''),
    hex.slice(4, 6).join(''),
    hex.slice(6, 8).join(''),
    hex.slice(8, 10).join(''),
    hex.slice(10).join(''),
  ].join('-');
}

export class SchedulesResource {
  constructor(private http: HttpClient) {}

  async create(
    request: CreateScheduleRequest,
    options?: RequestOptions,
  ): Promise<ScheduleResponse> {
    const res = await this.http.request<ScheduleResponse>({
      method: 'POST',
      path: '/schedules',
      body: {
        ...request,
        timezone: request.timezone ?? 'UTC',
        missed_run_policy: request.missed_run_policy ?? 'run_once',
      },
      headers: { 'Idempotency-Key': options?.idempotencyKey ?? newIdempotencyKey() },
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }

  async list(options?: RequestOptions): Promise<ListSchedulesResponse> {
    const res = await this.http.request<ListSchedulesResponse>({
      method: 'GET',
      path: '/schedules',
      signal: options?.signal,
    });
    return res.data;
  }

  async update(
    scheduleId: string,
    request: UpdateScheduleRequest,
    options?: RequestOptions,
  ): Promise<ScheduleResponse> {
    const res = await this.http.request<ScheduleResponse>({
      method: 'PATCH',
      path: `/schedules/${encodeURIComponent(scheduleId)}`,
      body: request,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
    return res.data;
  }

  async delete(scheduleId: string, options?: RequestOptions): Promise<void> {
    await this.http.request<unknown>({
      method: 'DELETE',
      path: `/schedules/${encodeURIComponent(scheduleId)}`,
      signal: options?.signal,
      idempotencyKey: options?.idempotencyKey,
    });
  }

  async listRuns(
    scheduleId: string,
    query?: { limit?: number; offset?: number },
    options?: { signal?: AbortSignal },
  ): Promise<ListScheduleRunsResponse> {
    const res = await this.http.request<ListScheduleRunsResponse>({
      method: 'GET',
      path: `/schedules/${encodeURIComponent(scheduleId)}/runs`,
      query,
      signal: options?.signal,
    });
    return res.data;
  }
}
