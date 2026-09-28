'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import { api, type FeeSchedule, type FeeCollectedSummary } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { BarChart3, ArrowRightLeft } from 'lucide-react';

export default function UsagePage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [fees, setFees] = useState<FeeSchedule | null>(null);
  const [collected, setCollected] = useState<FeeCollectedSummary | null>(null);
  const [totalTransactions, setTotalTransactions] = useState(0);
  const [totalVolume, setTotalVolume] = useState(0);

  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const [feeData, collectedData] = await Promise.all([
        api.getFeeSchedule().catch(() => null),
        api.listFeeCollected().catch(() => null),
      ]);
      setFees(feeData);
      setCollected(collectedData);

      let txCount = 0;
      let vol = 0;
      for (const id of walletIds) {
        try {
          const res = await api.listTransactions(id, 100);
          const txs = res.transactions || [];
          txCount += txs.length;
          vol += txs.reduce((sum, tx) => sum + parseFloat(tx.amount || '0'), 0);
        } catch {}
      }
      setTotalTransactions(txCount);
      setTotalVolume(vol);
    } catch {
      toast('Failed to load usage data', 'error');
    } finally {
      setLoading(false);
    }
  }, [walletIds, toast]);

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

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <Skeleton className="h-10 w-56" />
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          <Skeleton className="h-40" />
          <Skeleton className="h-40" />
        </div>
        <Skeleton className="h-48" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="Usage & Billing" description="Monitor your API usage and limits." />

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <ArrowRightLeft className="h-4 w-4" />
              Total Transactions
            </CardDescription>
            <CardTitle className="text-3xl">{totalTransactions}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Across {walletIds.length} wallet(s)</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center gap-2">
              <BarChart3 className="h-4 w-4" />
              Transfer Volume
            </CardDescription>
            <CardTitle className="text-3xl">{totalVolume.toFixed(7)}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">Total amount transferred</p>
          </CardContent>
        </Card>
      </div>

      {fees && (
        <Card>
          <CardHeader>
            <CardTitle>Fee Schedule</CardTitle>
            <CardDescription>Rates applied to your account.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Transfer Fee
                </p>
                <p className="text-2xl font-semibold">
                  {fees.transfer_fee_bps}{' '}
                  <span className="text-sm font-normal text-muted-foreground">bps</span>
                </p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Conversion Fee
                </p>
                <p className="text-2xl font-semibold">
                  {fees.conversion_fee_bps}{' '}
                  <span className="text-sm font-normal text-muted-foreground">bps</span>
                </p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Minimum Fee
                </p>
                <p className="text-2xl font-semibold">{fees.min_fee_amount}</p>
              </div>
              <div>
                <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  Fee Asset
                </p>
                <p className="text-2xl font-semibold">{fees.asset}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {collected?.collected && collected.collected.length > 0 && (
        <div className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">Collected Fees</h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            {collected.collected.map((item) => (
              <Card key={item.asset}>
                <CardContent className="p-5">
                  <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    {item.asset}
                  </p>
                  <p className="mt-1 text-2xl font-semibold">{item.total_fees}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {item.transfer_count} transfers
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
