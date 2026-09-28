export interface FluxaErrorBody {
  code: string;
  message: string;
  details?: unknown;
}

export class FluxaError extends Error {
  readonly statusCode: number;
  readonly code: string;
  readonly details?: unknown;

  constructor(statusCode: number, body: FluxaErrorBody) {
    super(body.message);
    this.name = 'FluxaError';
    this.statusCode = statusCode;
    this.code = body.code;
    this.details = body.details;
  }
}

export class AuthenticationError extends FluxaError {
  constructor(message = 'Invalid or missing API key') {
    super(401, { code: 'UNAUTHORIZED', message });
    this.name = 'AuthenticationError';
  }
}

export class NotFoundError extends FluxaError {
  constructor(message = 'Resource not found') {
    super(404, { code: 'NOT_FOUND', message });
    this.name = 'NotFoundError';
  }
}

export class ValidationError extends FluxaError {
  constructor(message: string, details?: unknown) {
    super(400, { code: 'VALIDATION_ERROR', message, details });
    this.name = 'ValidationError';
  }
}

export class RateLimitError extends FluxaError {
  retryAfter?: number;

  constructor(retryAfter?: number) {
    super(429, {
      code: 'RATE_LIMITED',
      message: `Rate limit exceeded${retryAfter ? `. Retry after ${retryAfter}s` : ''}`,
    });
    this.name = 'RateLimitError';
    this.retryAfter = retryAfter;
  }
}

export class ConflictError extends FluxaError {
  constructor(message: string) {
    super(409, { code: 'CONFLICT', message });
    this.name = 'ConflictError';
  }
}

export function classifyError(status: number, body: unknown): FluxaError {
  const parsed = body as FluxaErrorBody;

  if (typeof parsed?.code === 'string' && typeof parsed?.message === 'string') {
    switch (status) {
      case 400:
        return new ValidationError(parsed.message, parsed.details);
      case 401:
        return new AuthenticationError(parsed.message);
      case 404:
        return new NotFoundError(parsed.message);
      case 409:
        return new ConflictError(parsed.message);
      case 429:
        return new RateLimitError();
      default:
        return new FluxaError(status, parsed);
    }
  }

  return new FluxaError(status, {
    code: 'UNKNOWN_ERROR',
    message: `Request failed with status ${status}`,
  });
}
