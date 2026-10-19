import { describe, it, expect, vi, beforeEach } from 'vitest';
import { iterate, listAll, RepeatedCursorError } from '../src/pagination';
import type { Page, PaginationParams } from '../src/types/pagination';

interface MockItem {
  id: string;
  value: number;
}

const createMockPage = (
  items: MockItem[],
  hasNextPage: boolean,
  endCursor?: string
): Page<MockItem> => ({
  nodes: items,
  pageInfo: {
    hasNextPage,
    endCursor: endCursor ?? (items.length > 0 ? items[items.length - 1].id : null)
  }
});

const createFetchPage = (
  pages: Page<MockItem>[],
  delay: number = 0
) => {
  let callCount = 0;
  return async (params: PaginationParams, signal?: AbortSignal): Promise<Page<MockItem>> => {
    if (signal?.aborted) {
      throw new Error('Aborted');
    }
    if (delay > 0) {
      await new Promise(resolve => setTimeout(resolve, delay));
    }
    if (callCount >= pages.length) {
      throw new Error('Unexpected page request');
    }
    const page = pages[callCount++];
    return page;
  };
};

describe('pagination', () => {
  describe('iterate', () => {
    it('should handle empty result', async () => {
      const fetchPage = createFetchPage([
        createMockPage([], false)
      ]);
      const items: MockItem[] = [];
      for await (const item of iterate({ fetchPage })) {
        items.push(item);
      }
      expect(items).toEqual([]);
    });

    it('should handle single page', async () => {
      const items = [{ id: '1', value: 10 }];
      const fetchPage = createFetchPage([
        createMockPage(items, false)
      ]);
      const result: MockItem[] = [];
      for await (const item of iterate({ fetchPage })) {
        result.push(item);
      }
      expect(result).toEqual(items);
    });

    it('should handle multiple pages', async () => {
      const page1 = [{ id: '1', value: 10 }];
      const page2 = [{ id: '2', value: 20 }];
      const page3 = [{ id: '3', value: 30 }];
      const fetchPage = createFetchPage([
        createMockPage(page1, true, '1'),
        createMockPage(page2, true, '2'),
        createMockPage(page3, false, '3')
      ]);
      const result: MockItem[] = [];
      for await (const item of iterate({ fetchPage })) {
        result.push(item);
      }
      expect(result).toEqual([...page1, ...page2, ...page3]);
    });

    it('should support early exit', async () => {
      const fetchPage = createFetchPage([
        createMockPage([{ id: '1', value: 10 }], true, '1'),
        createMockPage([{ id: '2', value: 20 }], true, '2'),
        createMockPage([{ id: '3', value: 30 }], false, '3')
      ]);
      const result: MockItem[] = [];
      let count = 0;
      for await (const item of iterate({ fetchPage })) {
        result.push(item);
        count++;
        if (count >= 2) break;
      }
      expect(result.length).toBe(2);
      expect(result).toEqual([{ id: '1', value: 10 }, { id: '2', value: 20 }]);
    });

    it('should respect abort signal', async () => {
      const fetchPage = createFetchPage([
        createMockPage([{ id: '1', value: 10 }], true, '1'),
        createMockPage([{ id: '2', value: 20 }], false, '2')
      ], 10);

      const controller = new AbortController();
      const items: MockItem[] = [];

      setTimeout(() => controller.abort(), 5);

      try {
        for await (const item of iterate({ fetchPage, signal: controller.signal })) {
          items.push(item);
        }
        expect.fail('Should have thrown');
      } catch (error) {
        expect((error as Error).name).toBe('CursorPaginationError');
        expect((error as Error).message).toBe('Iteration aborted');
      }
      expect(items.length).toBe(0);
    });

    it('should detect repeated cursors', async () => {
      const fetchPage = createFetchPage([
        createMockPage([{ id: '1', value: 10 }], true, '1'),
        createMockPage([{ id: '1', value: 10 }], true, '1')
      ]);

      try {
        const items: MockItem[] = [];
        for await (const item of iterate({ fetchPage })) {
          items.push(item);
        }
        expect.fail('Should have thrown');
      } catch (error) {
        expect(error).toBeInstanceOf(RepeatedCursorError);
        expect((error as RepeatedCursorError).message).toContain('1');
      }
    });

    it('should preserve initial params', async () => {
      const fetchPage = vi.fn().mockImplementation((params: PaginationParams) => {
        return Promise.resolve(
          createMockPage([{ id: params.after || '1', value: 10 }], false, params.after)
        );
      });

      const initialParams = { first: 50, after: 'initial' };
      const items: MockItem[] = [];
      for await (const item of iterate({ fetchPage, initialParams })) {
        items.push(item);
      }

      expect(fetchPage).toHaveBeenCalledWith(
        expect.objectContaining({ first: 50, after: 'initial' }),
        undefined
      );
    });
  });

  describe('listAll', () => {
    it('should collect all items', async () => {
      const page1 = [{ id: '1', value: 10 }];
      const page2 = [{ id: '2', value: 20 }];
      const fetchPage = createFetchPage([
        createMockPage(page1, true, '1'),
        createMockPage(page2, false, '2')
      ]);
      const result = await listAll({ fetchPage });
      expect(result).toEqual([...page1, ...page2]);
    });

    it('should handle empty result', async () => {
      const fetchPage = createFetchPage([createMockPage([], false)]);
      const result = await listAll({ fetchPage });
      expect(result).toEqual([]);
    });
  });
});
