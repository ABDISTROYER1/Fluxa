'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import { api, type Transaction } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { ArrowRightLeft, ExternalLink, Plus, X } from 'lucide-react';

function statusBadge(status: string) {
  if (status === 'confirmed') return <Badge variant="success">{status}</Badge>;
  if (status === 'pending') return <Badge variant="warning">{status}</Badge>;
  return <Badge variant="danger">{status}</Badge>;
}

export default function TransfersPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const [transfers, setTransfers] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');
  const [showForm, setShowForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({
    from_wallet_id: '',
    to_wallet_id: '',
    asset: 'XLM',
    amount: '',
  });

  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  const fetchTransfers = useCallback(async () => {
    setLoading(true);
    try {
      const allTx: Transaction[] = [];
      for (const id of walletIds) {
        try {
          const res = await api.listTransactions(id, 50);
          allTx.push(...(res.transactions || []));
        } catch {}
      }
      const unique = Array.from(new Map(allTx.map((t) => [t.id, t])).values());
      unique.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
      setTransfers(unique);
    } catch {
      toast('Failed to load transfers', 'error');
    } finally {
      setLoading(false);
    }
  }, [walletIds, toast]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await fetchTransfers();
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchTransfers]);

  const handleCreateTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.createTransfer(form);
      toast('Transfer initiated', 'success');
      setShowForm(false);
      setForm({
        from_wallet_id: '',
        to_wallet_id: '',
        asset: 'XLM',
        amount: '',
      });
      await fetchTransfers();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Transfer failed', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const filtered = filter === 'all' ? transfers : transfers.filter((t) => t.status === filter);

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="Transfers" description="View and trace your transfer history.">
        <div className="flex items-center gap-3">
          <Select value={filter} onChange={(e) => setFilter(e.target.value)} className="w-40">
            <option value="all">All Statuses</option>
            <option value="confirmed">Confirmed</option>
            <option value="pending">Pending</option>
            <option value="failed">Failed</option>
          </Select>
          {walletIds.length >= 2 && (
            <Button
              variant={showForm ? 'secondary' : 'primary'}
              onClick={() => setShowForm(!showForm)}
            >
              {showForm ? (
                <>
                  <X className="h-4 w-4" /> Cancel
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4" /> New Transfer
                </>
              )}
            </Button>
          )}
        </div>
      </PageHeader>

      {showForm && (
        <Card className="max-w-2xl">
          <CardHeader>
            <CardTitle>New Transfer</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreateTransfer} className="flex flex-col gap-5">
              <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">From Wallet</label>
                  <Select
                    value={form.from_wallet_id}
                    onChange={(e) => setForm({ ...form, from_wallet_id: e.target.value })}
                    required
                  >
                    <option value="">Select wallet</option>
                    {walletIds.map((id) => (
                      <option key={id} value={id}>
                        {id.slice(0, 16)}...
                      </option>
                    ))}
                  </Select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">To Wallet</label>
                  <Select
                    value={form.to_wallet_id}
                    onChange={(e) => setForm({ ...form, to_wallet_id: e.target.value })}
                    required
                  >
                    <option value="">Select wallet</option>
                    {walletIds.map((id) => (
                      <option key={id} value={id}>
                        {id.slice(0, 16)}...
                      </option>
                    ))}
                  </Select>
                </div>
              </div>
              <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">Asset</label>
                  <Input
                    value={form.asset}
                    onChange={(e) => setForm({ ...form, asset: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">Amount</label>
                  <Input
                    value={form.amount}
                    onChange={(e) => setForm({ ...form, amount: e.target.value })}
                    required
                    placeholder="0.0000000"
                    className="font-mono"
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button
                  type="submit"
                  isLoading={submitting}
                  disabled={!form.from_wallet_id || !form.to_wallet_id || !form.amount}
                >
                  Initiate Transfer
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <Card>
        {filtered.length === 0 ? (
          <EmptyState
            icon={ArrowRightLeft}
            title="No transfers found"
            description="Create wallets and initiate a transfer to get started."
          />
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableHeader>ID</TableHeader>
                <TableHeader>Amount</TableHeader>
                <TableHeader>From / To</TableHeader>
                <TableHeader>Date</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader>Failure</TableHeader>
                <TableHeader>Stellar Tx</TableHeader>
              </TableRow>
            </TableHead>
            <TableBody>
              {filtered.map((tr) => (
                <TableRow key={tr.id}>
                  <TableCell className="font-mono text-muted-foreground">
                    {tr.id.slice(0, 8)}
                  </TableCell>
                  <TableCell className="font-medium">
                    {tr.amount} {tr.asset}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {tr.from_wallet_id.slice(0, 8)}... &rarr; {tr.to_wallet_id.slice(0, 8)}...
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(tr.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>{statusBadge(tr.status)}</TableCell>
                  <TableCell>{tr.failure_message || tr.failure_reason || '—'}</TableCell>
                  <TableCell>
                    {tr.tx_hash ? (
                      <a
                        href={`https://stellar.expert/explorer/public/tx/${tr.tx_hash}`}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary-hover hover:underline"
                      >
                        {tr.tx_hash.slice(0, 8)}...
                        <ExternalLink className="h-3.5 w-3.5" />
                      </a>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>
    </div>
  );
}
