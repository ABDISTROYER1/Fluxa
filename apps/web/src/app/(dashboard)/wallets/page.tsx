'use client';

import { useEffect, useState, useCallback } from 'react';
import { api, type WalletWithBalance, type WalletBalance } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Wallet, Copy, ExternalLink, Plus, Link2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Modal } from '@/components/ui/modal';

export default function WalletsPage() {
  const { getStoredWalletIds, addStoredWalletId } = useAuth();
  const { toast } = useToast();
  const [wallets, setWallets] = useState<WalletWithBalance[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [trustlineWallet, setTrustlineWallet] = useState<string | null>(null);
  const [trustlineForm, setTrustlineForm] = useState({ asset: 'USDC', issuer: '', limit: '' });
  const [trustlineLoading, setTrustlineLoading] = useState(false);

  const fetchWallets = useCallback(async () => {
    setLoading(true);
    const ids = getStoredWalletIds();
    const results = await Promise.all(
      ids.map(async (id) => {
        try {
          const res = await api.getWalletBalances(id);
          return {
            id,
            public_key: id,
            created_at: '',
            balances: res.balances,
          } as WalletWithBalance;
        } catch {
          return {
            id,
            public_key: id,
            created_at: '',
            balances: [] as WalletBalance[],
          } as WalletWithBalance;
        }
      }),
    );
    setWallets(results);
    setLoading(false);
  }, [getStoredWalletIds]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await fetchWallets();
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchWallets]);

  const handleCreateWallet = async () => {
    setCreating(true);
    try {
      const wallet = await api.createWallet();
      addStoredWalletId(wallet.id);
      toast('Wallet created successfully', 'success');
      await fetchWallets();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to create wallet', 'error');
    } finally {
      setCreating(false);
    }
  };

  const copyAddress = (addr: string) => {
    navigator.clipboard.writeText(addr);
    toast('Address copied', 'info');
  };

  const handleCreateTrustline = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!trustlineWallet) return;
    setTrustlineLoading(true);
    try {
      await api.createTrustline(trustlineWallet, {
        asset: trustlineForm.asset,
        issuer: trustlineForm.issuer || undefined,
        limit: trustlineForm.limit || undefined,
      });
      toast('Trustline submitted', 'success');
      setTrustlineWallet(null);
      await fetchWallets();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Trustline failed', 'error');
    } finally {
      setTrustlineLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          <Skeleton className="h-56" />
          <Skeleton className="h-56" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="Wallets" description="Manage your Stellar wallets and balances.">
        <Button onClick={handleCreateWallet} isLoading={creating}>
          <Plus className="h-4 w-4" />
          Create Wallet
        </Button>
      </PageHeader>

      {wallets.length === 0 ? (
        <EmptyState
          icon={Wallet}
          title="No wallets yet"
          description="Create your first wallet to start holding assets and making transfers."
          action={
            <Button onClick={handleCreateWallet} isLoading={creating}>
              <Plus className="h-4 w-4" />
              Create Wallet
            </Button>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
          {wallets.map((wallet) => (
            <Card key={wallet.id} className="flex flex-col">
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between">
                  <CardTitle>Wallet</CardTitle>
                  <Badge variant="success">Active</Badge>
                </div>
              </CardHeader>
              <CardContent className="flex flex-col gap-5">
                <div>
                  {wallet.balances.length === 0 ? (
                    <p className="text-3xl font-semibold tracking-tight text-muted-foreground">
                      0.00
                    </p>
                  ) : (
                    <div className="flex flex-col gap-1">
                      {wallet.balances.map((b) => (
                        <p key={b.asset_code} className="text-2xl font-semibold tracking-tight">
                          {parseFloat(b.balance).toFixed(7)}{' '}
                          <span className="text-sm font-normal text-muted-foreground">
                            {b.asset_code}
                          </span>
                        </p>
                      ))}
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between gap-3 rounded-lg border border-border bg-muted px-3 py-2">
                  <code className="font-mono text-xs text-muted-foreground break-all">
                    {wallet.id.slice(0, 16)}...
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-8 px-2"
                    onClick={() => copyAddress(wallet.id)}
                    aria-label="Copy wallet ID"
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                </div>

                <div className="flex items-center gap-3">
                  <a
                    href={`https://stellar.expert/explorer/public/account/${wallet.id}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary-hover hover:underline"
                  >
                    View on Stellar Expert
                    <ExternalLink className="h-3.5 w-3.5" />
                  </a>
                  <Button variant="ghost" size="sm" onClick={() => setTrustlineWallet(wallet.id)}>
                    <Link2 className="h-3.5 w-3.5" />
                    Trustline
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Modal
        open={!!trustlineWallet}
        onClose={() => setTrustlineWallet(null)}
        title="Add Trustline"
        description={`Enable ${trustlineWallet?.slice(0, 8)}... to hold a new asset.`}
      >
        <form onSubmit={handleCreateTrustline} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium">Asset Code</label>
            <Input
              value={trustlineForm.asset}
              onChange={(e) => setTrustlineForm({ ...trustlineForm, asset: e.target.value })}
              required
              placeholder="USDC"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium">Issuer Public Key</label>
            <Input
              value={trustlineForm.issuer}
              onChange={(e) => setTrustlineForm({ ...trustlineForm, issuer: e.target.value })}
              placeholder="GA... (leave empty for USDC/EURC defaults)"
              className="font-mono"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-sm font-medium">Limit (optional)</label>
            <Input
              value={trustlineForm.limit}
              onChange={(e) => setTrustlineForm({ ...trustlineForm, limit: e.target.value })}
              placeholder="100000"
              className="font-mono"
            />
          </div>
          <div className="flex justify-end gap-3">
            <Button type="button" variant="secondary" onClick={() => setTrustlineWallet(null)}>
              Cancel
            </Button>
            <Button type="submit" isLoading={trustlineLoading}>
              Submit
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
