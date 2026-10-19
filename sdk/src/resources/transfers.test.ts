import { describe, it, expect, vi } from 'vitest';
import { TransfersResource } from './transfers';
import { HttpClient, HttpResponse } from '../http';
import { ListTransactionsResponse, TransferResponse, CreateTransferRequest } from '../types';
import { RepeatedCursorError } from '../errors';

function mockTransfer(id: string, amount = '10.0000000'): TransferResponse {
  return {
    id,
    type: 'transfer',
    from_wallet_id: 'w_from',
    to_wallet_id: 'w_to',
    asset: 'USDC',
    amount,
    fee_amount: '0.0000100',
    net_amount: amount,
    fee_bps: 10,
    status: 'confirmed',
    created_at: '2026-09-30T12:00:00Z',
  };
}

describe('TransfersResource with pagination', () => {
  it('preserves existing list method signature and behavior', async () => {
    const mockResponse: HttpResponse<ListTransactionsResponse> = {
      status: 200,
      headers: new Headers(),
      data: {
        transactions: [mockTransfer('tx_1')],
      },
    };

    const mockHttp = {
      request: vi.fn(async () => mockResponse),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);
    const result = await resource.list({ wallet_id: 'w_1', limit: 10 });

    expect(result.transactions).toHaveLength(1);
    expect(result.transactions[0].id).toBe('tx_1');
    expect(mockHttp.request).toHaveBeenCalledWith({
      method: 'GET',
      path: '/transfers',
      query: { wallet_id: 'w_1', limit: 10 },
      signal: undefined,
    });
  });

  it('listPage returns standardized Page<TransferResponse>', async () => {
    const mockResponse: HttpResponse<ListTransactionsResponse> = {
      status: 200,
      headers: new Headers(),
      data: {
        transactions: [mockTransfer('tx_1'), mockTransfer('tx_2')],
        next_cursor: 'cur_page_2',
      },
    };

    const mockHttp = {
      request: vi.fn(async () => mockResponse),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);
    const page = await resource.listPage({ wallet_id: 'w_1' });

    expect(page.items).toHaveLength(2);
    expect(page.items[0].id).toBe('tx_1');
    expect(page.nextCursor).toBe('cur_page_2');
    expect(page.hasNextPage).toBe(true);
  });

  it('iterate iterates through multiple pages with for await...of', async () => {
    let callCount = 0;
    const mockHttp = {
      request: vi.fn(async ({ query }: { query?: { cursor?: string } }) => {
        callCount++;
        if (query?.cursor === 'c1') {
          return {
            status: 200,
            headers: new Headers(),
            data: {
              transactions: [mockTransfer('tx_3')],
              next_cursor: null,
            },
          };
        }
        return {
          status: 200,
          headers: new Headers(),
          data: {
            transactions: [mockTransfer('tx_1'), mockTransfer('tx_2')],
            next_cursor: 'c1',
          },
        };
      }),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);
    const txIds: string[] = [];

    for await (const tx of resource.iterate({ wallet_id: 'w_1', limit: 2 })) {
      txIds.push(tx.id);
    }

    expect(txIds).toEqual(['tx_1', 'tx_2', 'tx_3']);
    expect(callCount).toBe(2);
  });

  it('listAll fetches all transfers across multiple pages', async () => {
    const mockHttp = {
      request: vi.fn(async ({ query }: { query?: { cursor?: string } }) => {
        if (query?.cursor === 'cursor_2') {
          return {
            status: 200,
            headers: new Headers(),
            data: {
              transactions: [mockTransfer('tx_final')],
              next_cursor: null,
            },
          };
        }
        return {
          status: 200,
          headers: new Headers(),
          data: {
            transactions: [mockTransfer('tx_first')],
            next_cursor: 'cursor_2',
          },
        };
      }),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);
    const all = await resource.listAll({ wallet_id: 'w_1' });

    expect(all).toHaveLength(2);
    expect(all[0].id).toBe('tx_first');
    expect(all[1].id).toBe('tx_final');
  });

  it('throws RepeatedCursorError if next_cursor forms a loop', async () => {
    const mockHttp = {
      request: vi.fn(async () => ({
        status: 200,
        headers: new Headers(),
        data: {
          transactions: [mockTransfer('tx_loop')],
          next_cursor: 'duplicate_cursor',
        },
      })),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);

    await expect(async () => {
      await resource.listAll({ wallet_id: 'w_1' });
    }).rejects.toThrow(RepeatedCursorError);
  });

  it('create and get methods continue to work as expected', async () => {
    const mockHttp = {
      request: vi.fn(async () => ({
        status: 200,
        headers: new Headers(),
        data: mockTransfer('tx_created'),
      })),
    } as unknown as HttpClient;

    const resource = new TransfersResource(mockHttp);
    const req: CreateTransferRequest = {
      from_wallet_id: 'w1',
      to_wallet_id: 'w2',
      asset: 'USDC',
      amount: '50.0000000',
    };

    const created = await resource.create(req);
    expect(created.id).toBe('tx_created');

    const fetched = await resource.get('tx_created');
    expect(fetched.id).toBe('tx_created');
  });
});
