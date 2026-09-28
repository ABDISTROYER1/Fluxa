'use client';

import { useEffect, useState, useCallback } from 'react';
import {
  api,
  type HealthResponse,
  type FeeSchedule,
  type Transaction,
  type WalletBalance,
} from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
import { Activity, Wallet, ArrowRightLeft } from 'lucide-react';

interface WalletData {
  id: string;
  balances: WalletBalance[];
}

function statusBadge(status?: string) {
  if (status === 'ok') return <Badge variant="success">Healthy</Badge>;
  if (status === 'degraded') return <Badge variant="warning">Degraded</Badge>;
  return <Badge variant="danger">Down</Badge>;
}

function txStatusBadge(status: string) {
  if (status === 'confirmed') return <Badge variant="success">{status}</Badge>;
  if (status === 'pending') return <Badge variant="warning">{status}</Badge>;
  return <Badge variant="danger">{status}</Badge>;
}

export default function OverviewPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [fees, setFees] = useState<FeeSchedule | null>(null);
  const [wallets, setWallets] = useState<WalletData[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [healthData, feeData] = await Promise.all([
        api.getHealth().catch(() => null),
        api.getFeeSchedule().catch(() => null),
      ]);
      setHealth(healthData);
      setFees(feeData);

      const walletIds = getStoredWalletIds();
      const walletResults = await Promise.all(
        walletIds.map(async (id) => {
          try {
            const res = await api.getWalletBalances(id);
            return { id, balances: res.balances };
          } catch {
            return { id, balances: [] as WalletBalance[] };
          }
        }),
      );
      setWallets(walletResults);

      if (walletIds.length > 0) {
        try {
          const txRes = await api.listTransactions(walletIds[0], 10);
          setTransactions(txRes.transactions || []);
        } catch {
          setTransactions([]);
        }
      }
    } catch {
      toast('Failed to load overview data', 'error');
    } finally {
      setLoading(false);
    }
  }, [getStoredWalletIds, toast]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await fetchData();
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchData]);

  const totalBalance = wallets.reduce((sum, w) => {
    return sum + w.balances.reduce((s, b) => s + parseFloat(b.balance || '0'), 0);
  }, 0);

  const totalTransferVolume = transactions.reduce((sum, tx) => {
    return sum + parseFloat(tx.amount || '0');
  }, 0);

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <Skeleton className="h-10 w-64" />
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Skeleton className="h-40" />
          <Skeleton className="h-40" />
          <Skeleton className="h-40" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Overview"
        description={
          health?.status === 'ok'
            ? 'All systems operational.'
            : health
              ? `System status: ${health.status}`
              : 'Could not reach API'
        }
      />

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <Wallet className="h-4 w-4" />
              Wallets
            </CardDescription>
            <CardTitle className="text-3xl">{wallets.length}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              {totalBalance > 0
                ? `${totalBalance.toFixed(7)} total across all assets`
                : 'No balances loaded'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <ArrowRightLeft className="h-4 w-4" />
              Transfer Volume
            </CardDescription>
            <CardTitle className="text-3xl">
              {totalTransferVolume > 0 ? totalTransferVolume.toFixed(7) : '0'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Transfer fee: {fees ? `${fees.transfer_fee_bps} bps` : 'N/A'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              API Health
            </CardDescription>
            <CardTitle className="text-3xl">{statusBadge(health?.status)}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2">
              {health?.services &&
                Object.entries(health.services).map(([name, status]) => (
                  <Badge key={name} variant={status === 'up' ? 'success' : 'danger'}>
                    {name}: {status}
                  </Badge>
                ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {fees && (
        <Card>
          <CardHeader>
            <CardTitle>Fee Schedule</CardTitle>
            <CardDescription>Rates applied to transfers and conversions.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Transfer Fee
                </p>
                <p className="text-lg font-semibold">{fees.transfer_fee_bps} bps</p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Conversion Fee
                </p>
                <p className="text-lg font-semibold">{fees.conversion_fee_bps} bps</p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Min Fee
                </p>
                <p className="text-lg font-semibold">{fees.min_fee_amount}</p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Asset
                </p>
                <p className="text-lg font-semibold">{fees.asset}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">Recent Transactions</h2>
        <Card>
          {transactions.length === 0 ? (
            <EmptyState
              title="No transactions yet"
              description="Create a wallet and initiate a transfer to get started."
            />
          ) : (
            <Table>
              <TableHead>
                <TableRow>
                  <TableHeader>ID</TableHeader>
                  <TableHeader>Type</TableHeader>
                  <TableHeader>Amount</TableHeader>
                  <TableHeader>Date</TableHeader>
                  <TableHeader>Status</TableHeader>
                </TableRow>
              </TableHead>
              <TableBody>
                {transactions.map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell className="font-mono text-muted-foreground">
                      {tx.id.slice(0, 8)}
                    </TableCell>
                    <TableCell>{tx.type}</TableCell>
                    <TableCell className="font-medium">
                      {tx.amount} {tx.asset}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(tx.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>{txStatusBadge(tx.status)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </Card>
      </div>
    </div>
  );
}
