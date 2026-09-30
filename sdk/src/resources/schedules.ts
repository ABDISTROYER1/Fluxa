import { HttpClient, RequestOptions } from '../http';
import {
  CreateScheduleRequest,
  UpdateScheduleRequest,
  ScheduleResponse,
  ListSchedulesResponse,
} from '../types';

export class SchedulesResource {
  constructor(private http: HttpClient) {}

  async create(
    request: CreateScheduleRequest,
    options?: RequestOptions,
  ): Promise<ScheduleResponse> {
    const res = await this.http.request<ScheduleResponse>({
      method: 'POST',
      path: '/schedules',
      body: request,
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
}
