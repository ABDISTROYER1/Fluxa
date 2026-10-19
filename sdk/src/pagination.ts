import { RequestOptions } from './http';
import { RepeatedCursorError } from './errors';

/**
 * Standardized cursor-paginated page container.
 */
export interface Page<T> {
  /** Items returned for the current page */
  items: T[];
  /** Next cursor token, or null if no further pages */
  nextCursor: string | null;
  /** Whether additional pages are available */
  hasNextPage: boolean;
  /** Alias for nextCursor matching snake_case API conventions */
  next_cursor?: string | null;
}

/**
 * Page fetcher signature for paginated list endpoints.
 */
export type PageFetcher<T, Q = Record<string, unknown>> = (
  query: Q,
  options?: RequestOptions,
) => Promise<Page<T>>;

/**
 * Configuration options for cursor pagination helpers.
 */
export interface PaginateOptions<T, Q = Record<string, unknown>> {
  /** Function to fetch a single page */
  fetchPage: PageFetcher<T, Q>;
  /** Initial query parameters (filters, limits, sorting) */
  query?: Q;
  /** HTTP request options (e.g. AbortSignal, headers) */
  options?: RequestOptions;
  /** Name of the query parameter used for cursor (defaults to 'cursor') */
  cursorParam?: string;
}

/**
 * Helper to construct a typed Page object from items and next-cursor metadata.
 */
export function createPage<T>(items: T[], nextCursor?: string | null): Page<T> {
  const cursor =
    typeof nextCursor === 'string' && nextCursor.trim().length > 0 ? nextCursor.trim() : null;
  return {
    items: Array.isArray(items) ? items : [],
    nextCursor: cursor,
    next_cursor: cursor,
    hasNextPage: cursor !== null,
  };
}

/**
 * Async generator that iterates item-by-item across all cursor-paginated pages.
 *
 * Guarantees:
 * - Lazy evaluation: Breaking out of iteration stops further HTTP requests immediately.
 * - AbortSignal forwarding: Checks signal before every request and during item yielding.
 * - State preservation: Preserves all original query parameters (filters, sort, limit).
 * - Loop safety: Detects repeated cursors and throws RepeatedCursorError.
 */
export async function* paginate<T, Q extends Record<string, unknown> = Record<string, unknown>>(
  config: PaginateOptions<T, Q>,
): AsyncIterableIterator<T> {
  const { fetchPage, query = {} as Q, options, cursorParam = 'cursor' } = config;

  const visitedCursors = new Set<string>();
  let currentCursor: string | null = (query[cursorParam] as string | undefined) ?? null;
  if (currentCursor) {
    visitedCursors.add(currentCursor);
  }

  let hasNext = true;
  let isFirstPage = true;

  while (hasNext) {
    // Check if aborted before initiating request
    if (options?.signal?.aborted) {
      if (typeof options.signal.throwIfAborted === 'function') {
        options.signal.throwIfAborted();
      }
      const abortErr = new Error('The operation was aborted');
      abortErr.name = 'AbortError';
      throw abortErr;
    }

    // Preserve original query and attach next cursor
    const pageQuery: Q = isFirstPage
      ? { ...query }
      : ({ ...query, [cursorParam]: currentCursor } as Q);

    // Fetch the page with forwarded request options
    const page = await fetchPage(pageQuery, options);

    // Yield items from current page
    if (page.items && page.items.length > 0) {
      for (const item of page.items) {
        if (options?.signal?.aborted) {
          if (typeof options.signal.throwIfAborted === 'function') {
            options.signal.throwIfAborted();
          }
          const abortErr = new Error('The operation was aborted');
          abortErr.name = 'AbortError';
          throw abortErr;
        }
        yield item;
      }
    }

    const nextCursor = page.nextCursor ?? page.next_cursor ?? null;

    if (!nextCursor || !page.hasNextPage) {
      hasNext = false;
      break;
    }

    // Guard against malformed server looping
    if (visitedCursors.has(nextCursor)) {
      throw new RepeatedCursorError(nextCursor);
    }

    visitedCursors.add(nextCursor);
    currentCursor = nextCursor;
    isFirstPage = false;
  }
}

/**
 * Async generator that iterates page-by-page across all cursor-paginated pages.
 */
export async function* paginatePages<
  T,
  Q extends Record<string, unknown> = Record<string, unknown>,
>(config: PaginateOptions<T, Q>): AsyncIterableIterator<Page<T>> {
  const { fetchPage, query = {} as Q, options, cursorParam = 'cursor' } = config;

  const visitedCursors = new Set<string>();
  let currentCursor: string | null = (query[cursorParam] as string | undefined) ?? null;
  if (currentCursor) {
    visitedCursors.add(currentCursor);
  }

  let hasNext = true;
  let isFirstPage = true;

  while (hasNext) {
    if (options?.signal?.aborted) {
      if (typeof options.signal.throwIfAborted === 'function') {
        options.signal.throwIfAborted();
      }
      const abortErr = new Error('The operation was aborted');
      abortErr.name = 'AbortError';
      throw abortErr;
    }

    const pageQuery: Q = isFirstPage
      ? { ...query }
      : ({ ...query, [cursorParam]: currentCursor } as Q);

    const page = await fetchPage(pageQuery, options);
    yield page;

    const nextCursor = page.nextCursor ?? page.next_cursor ?? null;
    if (!nextCursor || !page.hasNextPage) {
      hasNext = false;
      break;
    }

    if (visitedCursors.has(nextCursor)) {
      throw new RepeatedCursorError(nextCursor);
    }

    visitedCursors.add(nextCursor);
    currentCursor = nextCursor;
    isFirstPage = false;
  }
}

/**
 * Fetches all items across all pages and returns them as a single consolidated array.
 */
export async function paginateAll<T, Q extends Record<string, unknown> = Record<string, unknown>>(
  config: PaginateOptions<T, Q>,
): Promise<T[]> {
  const items: T[] = [];
  for await (const item of paginate(config)) {
    items.push(item);
  }
  return items;
}
