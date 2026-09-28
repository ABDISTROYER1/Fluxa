import { classifyError, FluxaError, RateLimitError } from './errors';

export interface HttpClientConfig {
  baseUrl: string;
  apiKey: string;
  timeout: number;
  maxRetries: number;
  retryDelay: number;
}

export interface RequestOptions {
  method: string;
  path: string;
  body?: unknown;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  query?: Record<string, any>;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

export interface HttpResponse<T> {
  data: T;
  status: number;
  headers: Headers;
}

function buildQueryString(
  params?:
    Record<string, string | number | undefined> | { [key: string]: string | number | undefined },
): string {
  if (!params) return '';
  const entries = Object.entries(params).filter(([, v]) => v !== undefined && v !== null);
  if (entries.length === 0) return '';
  const qs = new URLSearchParams();
  for (const [k, v] of entries) {
    qs.set(k, String(v));
  }
  return `?${qs.toString()}`;
}

function isRetryable(status: number): boolean {
  return status === 429 || status >= 500;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class HttpClient {
  private config: HttpClientConfig;

  constructor(config: HttpClientConfig) {
    this.config = config;
  }

  async request<T>(options: RequestOptions): Promise<HttpResponse<T>> {
    const { method, path, body, query, headers: extraHeaders, signal } = options;
    const url = `${this.config.baseUrl}/v1${path}${buildQueryString(query)}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${this.config.apiKey}`,
      ...extraHeaders,
    };

    let lastError: Error | undefined;

    for (let attempt = 0; attempt <= this.config.maxRetries; attempt++) {
      if (attempt > 0) {
        const delay = this.config.retryDelay * Math.pow(2, attempt - 1);
        await sleep(delay);
      }

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

      // Combine external signal with timeout signal
      if (signal) {
        if (signal.aborted) {
          controller.abort();
        } else {
          signal.addEventListener('abort', () => controller.abort(), {
            once: true,
          });
        }
      }

      try {
        const res = await fetch(url, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

        // Handle empty responses (204 No Content)
        if (res.status === 204) {
          return { data: undefined as unknown as T, status: res.status, headers: res.headers };
        }

        const contentType = res.headers.get('content-type') ?? '';
        let responseData: unknown;

        if (contentType.includes('application/json')) {
          responseData = await res.json();
        } else if (contentType.includes('text/csv')) {
          responseData = await res.text();
        } else {
          responseData = await res.text();
        }

        if (!res.ok) {
          const error = classifyError(res.status, responseData);

          if (error instanceof RateLimitError) {
            const retryHeader = res.headers.get('retry-after');
            if (retryHeader) {
              error.retryAfter = parseInt(retryHeader, 10);
            }
          }

          if (isRetryable(res.status) && attempt < this.config.maxRetries) {
            lastError = error;
            const retryDelay =
              error instanceof RateLimitError && error.retryAfter
                ? error.retryAfter * 1000
                : undefined;
            if (retryDelay) {
              await sleep(retryDelay);
            }
            continue;
          }

          throw error;
        }

        return { data: responseData as T, status: res.status, headers: res.headers };
      } catch (err) {
        clearTimeout(timeoutId);

        if (err instanceof FluxaError) {
          throw err;
        }

        // Network or abort errors are retryable
        if (attempt < this.config.maxRetries) {
          lastError = err instanceof Error ? err : new Error(String(err));
          continue;
        }

        if (err instanceof DOMException && err.name === 'AbortError') {
          throw new FluxaError(408, {
            code: 'TIMEOUT',
            message: `Request timed out after ${this.config.timeout}ms`,
          });
        }

        throw new FluxaError(0, {
          code: 'NETWORK_ERROR',
          message: lastError?.message ?? 'Network request failed',
        });
      }
    }

    throw lastError ?? new FluxaError(0, { code: 'UNKNOWN', message: 'Request failed' });
  }
}
