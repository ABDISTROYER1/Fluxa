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
    this.name = "FluxaError";
    this.statusCode = statusCode;
    this.code = body.code;
    this.details = body.details;
  }
}

export class AuthenticationError extends FluxaError {
  constructor(body: FluxaErrorBody) {
    super(401, body);
    this.name = "AuthenticationError";
  }
}

export class NotFoundError extends FluxaError {
  constructor(body: FluxaErrorBody) {
    super(404, body);
    this.name = "NotFoundError";
  }
}

export class ValidationError extends FluxaError {
  constructor(body: FluxaErrorBody) {
    super(400, body);
    this.name = "ValidationError";
  }
}

export class RateLimitError extends FluxaError {
  retryAfter?: number;

  constructor(body: FluxaErrorBody, retryAfter?: number) {
    super(429, body);
    this.name = "RateLimitError";
    this.retryAfter = retryAfter;
  }
}

export class ConflictError extends FluxaError {
  constructor(body: FluxaErrorBody) {
    super(409, body);
    this.name = "ConflictError";
  }
}

export function classifyError(status: number, body: unknown): FluxaError {
  // The API wraps errors as { "error": { "code": "...", "message": "..." } }
  // with an optional top-level "validation_errors" array for 400s.
  const envelope = body as { error?: FluxaErrorBody; validation_errors?: unknown };
  const detail = envelope?.error;

  if (typeof detail?.code === "string" && typeof detail?.message === "string") {
    const parsed: FluxaErrorBody = {
      code: detail.code,
      message: detail.message,
      details: envelope.validation_errors ?? detail.details,
    };

    switch (status) {
      case 400:
        return new ValidationError(parsed);
      case 401:
        return new AuthenticationError(parsed);
      case 404:
        return new NotFoundError(parsed);
      case 409:
        return new ConflictError(parsed);
      case 422:
        return new FluxaError(status, parsed);
      case 429:
        return new RateLimitError(parsed);
      default:
        return new FluxaError(status, parsed);
    }
  }

  return new FluxaError(status, {
    code: "UNKNOWN_ERROR",
    message: `Request failed with status ${status}`,
  });
}
