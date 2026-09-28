import { HttpClient } from '../http';
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
    options?: { signal?: AbortSignal },
  ): Promise<ScheduleResponse> {
    const res = await this.http.request<ScheduleResponse>({
      method: 'POST',
      path: '/schedules',
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async list(options?: { signal?: AbortSignal }): Promise<ListSchedulesResponse> {
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
    options?: { signal?: AbortSignal },
  ): Promise<ScheduleResponse> {
    const res = await this.http.request<ScheduleResponse>({
      method: 'PATCH',
      path: `/schedules/${encodeURIComponent(scheduleId)}`,
      body: request,
      signal: options?.signal,
    });
    return res.data;
  }

  async delete(scheduleId: string, options?: { signal?: AbortSignal }): Promise<void> {
    await this.http.request<unknown>({
      method: 'DELETE',
      path: `/schedules/${encodeURIComponent(scheduleId)}`,
      signal: options?.signal,
    });
  }
}
