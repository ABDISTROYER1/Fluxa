import {
  Page,
  PageInfo,
  PaginationParams,
  IterationOptions,
  RepeatedCursorError,
  CursorPaginationError
} from './types/pagination';

const DEFAULT_PAGE_SIZE = 100;

export async function* iterate<T>(
  options: IterationOptions<T>
): AsyncIterable<T> {
  const { fetchPage, initialParams = {}, signal } = options;
  const visitedCursors = new Set<string>();
  let currentParams: PaginationParams = { ...initialParams };
  let hasNextPage = true;
  let page: Page<T> | null = null;

  try {
    while (hasNextPage) {
      if (signal?.aborted) {
        throw new CursorPaginationError('Iteration aborted');
      }

      page = await fetchPage(currentParams, signal);
      const { nodes, pageInfo } = page;

      if (pageInfo.endCursor) {
        if (visitedCursors.has(pageInfo.endCursor)) {
          throw new RepeatedCursorError(pageInfo.endCursor);
        }
        visitedCursors.add(pageInfo.endCursor);
      }

      yield* nodes;

      hasNextPage = pageInfo.hasNextPage;
      currentParams = {
        ...currentParams,
        first: currentParams.first ?? DEFAULT_PAGE_SIZE,
        after: pageInfo.endCursor ?? undefined
      };

      if (!hasNextPage) {
        break;
      }
    }
  } finally {
    if (page?.pageInfo.endCursor && signal?.aborted) {
      signal.throwIfAborted();
    }
  }
}

export async function listAll<T>(
  options: IterationOptions<T>
): Promise<T[]> {
  const results: T[] = [];
  for await (const item of iterate(options)) {
    results.push(item);
  }
  return results;
}

export function buildPaginationParams(
  size?: number,
  cursor?: string
): PaginationParams {
  return {
    first: size ?? DEFAULT_PAGE_SIZE,
    after: cursor
  };
}

export function hasNextPage(pageInfo: PageInfo): boolean {
  return pageInfo.hasNextPage && pageInfo.endCursor !== null;
}
