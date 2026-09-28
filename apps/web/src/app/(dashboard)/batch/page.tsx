'use client';

import { useState, useMemo } from 'react';
import { api, type BatchResponse } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Layers, Plus, Trash2, Download } from 'lucide-react';

interface BatchItem {
  to_wallet_id: string;
  asset: string;
  amount: string;
  reference: string;
}

export default function BatchPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  const [fromWalletId, setFromWalletId] = useState('');
  const [items, setItems] = useState<BatchItem[]>([
    { to_wallet_id: '', asset: 'XLM', amount: '', reference: '' },
  ]);
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<BatchResponse | null>(null);
  const [lookupId, setLookupId] = useState('');
  const [lookupLoading, setLookupLoading] = useState(false);

  const addItem = () =>
    setItems((p) => [...p, { to_wallet_id: '', asset: 'XLM', amount: '', reference: '' }]);
  const removeItem = (idx: number) => setItems((p) => p.filter((_, i) => i !== idx));
  const updateItem = (idx: number, field: keyof BatchItem, val: string) =>
    setItems((p) => p.map((it, i) => (i === idx ? { ...it, [field]: val } : it)));

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const res = await api.createBatch({
        from_wallet_id: fromWalletId,
        transfers: items,
      });
      setResult(res);
      toast(`Batch ${res.status} — ${res.total_count} transfers`, 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Batch failed', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleLookup = async () => {
    if (!lookupId) return;
    setLookupLoading(true);
    try {
      const r = await api.getBatch(lookupId);
      setResult(r);
      toast('Batch fetched', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Not found', 'error');
    } finally {
      setLookupLoading(false);
    }
  };

  const handleExport = async () => {
    if (!result) return;
    try {
      const csv = await api.exportBatchCsv(result.id);
      const blob = new Blob([csv], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `batch-${result.id}.csv`;
      a.click();
      URL.revokeObjectURL(url);
      toast('CSV downloaded', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Export failed', 'error');
    }
  };

  const statusBadge = (s: string) => {
    if (s === 'completed') return <Badge variant="success">{s}</Badge>;
    if (s === 'partial') return <Badge variant="warning">{s}</Badge>;
    if (s === 'failed') return <Badge variant="danger">{s}</Badge>;
    return <Badge variant="default">{s}</Badge>;
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Batch Transfers"
        description="Send up to 100 transfers in one call. CSV export included."
      >
        <div className="flex items-center gap-2">
          <Input
            value={lookupId}
            onChange={(e) => setLookupId(e.target.value)}
            placeholder="Batch ID"
            className="w-64"
          />
          <Button variant="secondary" onClick={handleLookup} isLoading={lookupLoading}>
            Fetch
          </Button>
        </div>
      </PageHeader>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Layers className="h-5 w-5" />
            Create Batch
          </CardTitle>
          <CardDescription>From one wallet to many recipients.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="flex flex-col gap-5">
            <div className="flex flex-col gap-1.5 max-w-xs">
              <label className="text-sm font-medium">From Wallet</label>
              <Select
                value={fromWalletId}
                onChange={(e) => setFromWalletId(e.target.value)}
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

            <div className="flex flex-col gap-3">
              {items.map((it, idx) => (
                <div
                  key={idx}
                  className="grid grid-cols-12 gap-2 items-end rounded-lg border border-border p-3"
                >
                  <div className="col-span-5 flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">To Wallet</label>
                    <Select
                      value={it.to_wallet_id}
                      onChange={(e) => updateItem(idx, 'to_wallet_id', e.target.value)}
                      required
                    >
                      <option value="">Select</option>
                      {walletIds.map((id) => (
                        <option key={id} value={id}>
                          {id.slice(0, 16)}...
                        </option>
                      ))}
                    </Select>
                  </div>
                  <div className="col-span-2 flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Asset</label>
                    <Input
                      value={it.asset}
                      onChange={(e) => updateItem(idx, 'asset', e.target.value)}
                      required
                    />
                  </div>
                  <div className="col-span-2 flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Amount</label>
                    <Input
                      value={it.amount}
                      onChange={(e) => updateItem(idx, 'amount', e.target.value)}
                      required
                      className="font-mono"
                    />
                  </div>
                  <div className="col-span-2 flex flex-col gap-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Reference</label>
                    <Input
                      value={it.reference}
                      onChange={(e) => updateItem(idx, 'reference', e.target.value)}
                    />
                  </div>
                  <div className="col-span-1">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeItem(idx)}
                      disabled={items.length === 1}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>

            <div className="flex gap-3">
              <Button type="button" variant="secondary" onClick={addItem}>
                <Plus className="h-4 w-4" />
                Add Recipient
              </Button>
              <Button
                type="submit"
                isLoading={submitting}
                disabled={!fromWalletId || items.length === 0}
              >
                Submit Batch ({items.length})
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {result && (
        <Card>
          <CardHeader className="flex-row items-center justify-between">
            <div>
              <CardTitle>Batch {result.id.slice(0, 8)}</CardTitle>
              <CardDescription>
                {result.total_count} total · {result.success_count} success · {result.failed_count}{' '}
                failed
              </CardDescription>
            </div>
            <div className="flex items-center gap-3">
              {statusBadge(result.status)}
              <Button variant="secondary" size="sm" onClick={handleExport}>
                <Download className="h-4 w-4" />
                Export CSV
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {result.transfers && result.transfers.length > 0 ? (
              <Table>
                <TableHead>
                  <TableRow>
                    <TableHeader>ID</TableHeader>
                    <TableHeader>To</TableHeader>
                    <TableHeader>Amount</TableHeader>
                    <TableHeader>Status</TableHeader>
                    <TableHeader>Failure</TableHeader>
                    <TableHeader>Tx Hash</TableHeader>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {result.transfers.map((t) => (
                    <TableRow key={t.id}>
                      <TableCell className="font-mono text-xs">{t.id.slice(0, 8)}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {t.to_wallet_id.slice(0, 8)}...
                      </TableCell>
                      <TableCell>
                        {t.amount} {t.asset}
                      </TableCell>
                      <TableCell>{statusBadge(t.status)}</TableCell>
                      <TableCell>{t.failure_message || t.failure_reason || '—'}</TableCell>
                      <TableCell className="font-mono text-xs">
                        {t.tx_hash?.slice(0, 8) || '—'}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <p className="text-sm text-muted-foreground">
                No per-transfer details yet — poll again shortly.
              </p>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
