export interface PageInfo {
  hasNextPage: boolean;
  endCursor: string | null;
}

export interface Page<T> {
  nodes: T[];
  pageInfo: PageInfo;
}

export interface PaginationParams {
  first?: number;
  after?: string;
}

export interface IterationOptions<T> {
  fetchPage: (params: PaginationParams, signal?: AbortSignal) => Promise<Page<T>>;
  initialParams?: PaginationParams;
  signal?: AbortSignal;
}

export class CursorPaginationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'CursorPaginationError';
  }
}

export class RepeatedCursorError extends CursorPaginationError {
  constructor(cursor: string) {
    super(`Repeated cursor detected: ${cursor}`);
    this.name = 'RepeatedCursorError';
  }
}
