'use client';

import { useState, useMemo } from 'react';
import { api, type QuoteResponse } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Coins, ArrowRightLeft } from 'lucide-react';
import type { RateResponse } from '@/lib/api';

export default function ConversionsPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  const [fromAsset, setFromAsset] = useState('USDC');
  const [toAsset, setToAsset] = useState('XLM');
  const [amount, setAmount] = useState('');
  const [walletId, setWalletId] = useState('');
  const [quote, setQuote] = useState<QuoteResponse | null>(null);
  const [rate, setRate] = useState<RateResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isConverting, setIsConverting] = useState(false);

  const handleGetQuote = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    try {
      const q = await api.getQuote({ from_asset: fromAsset, to_asset: toAsset, amount });
      setQuote(q);
      toast(
        `Quote ${q.id.slice(0, 8)} — expires ${new Date(q.expires_at).toLocaleTimeString()}`,
        'success',
      );
      // also fetch rate for display
      try {
        const r = await api.getRates(fromAsset, toAsset);
        setRate(r);
      } catch {}
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to get quote', 'error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleConvert = async () => {
    if (!quote || !walletId) return;
    setIsConverting(true);
    try {
      const conv = await api.convert({ wallet_id: walletId, quote_id: quote.id });
      toast(
        `Converted ${conv.source_amount} ${conv.source_asset} → ${conv.dest_amount} ${conv.dest_asset}`,
        'success',
      );
      setQuote(null);
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Conversion failed', 'error');
    } finally {
      setIsConverting(false);
    }
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Conversions"
        description="Convert between assets with real-time FX rates (alias of FX)."
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ArrowRightLeft className="h-5 w-5" />
              New Conversion
            </CardTitle>
            <CardDescription>
              Quotes live for 30s — convert with a wallet and quote ID.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <form onSubmit={handleGetQuote} className="flex flex-col gap-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">From Asset</label>
                  <Select value={fromAsset} onChange={(e) => setFromAsset(e.target.value)}>
                    <option value="USDC">USDC</option>
                    <option value="EURC">EURC</option>
                    <option value="XLM">XLM</option>
                  </Select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">To Asset</label>
                  <Select value={toAsset} onChange={(e) => setToAsset(e.target.value)}>
                    <option value="EURC">EURC</option>
                    <option value="USDC">USDC</option>
                    <option value="XLM">XLM</option>
                  </Select>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Amount</label>
                <Input
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                  placeholder="100.00"
                  className="font-mono"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Wallet ID</label>
                <Select value={walletId} onChange={(e) => setWalletId(e.target.value)} required>
                  <option value="">Select wallet</option>
                  {walletIds.map((id) => (
                    <option key={id} value={id}>
                      {id.slice(0, 16)}...
                    </option>
                  ))}
                </Select>
              </div>
              <Button type="submit" isLoading={isLoading}>
                Get Quote
              </Button>
            </form>
          </CardContent>
        </Card>

        <div className="flex flex-col gap-6">
          {quote && (
            <Card>
              <CardHeader>
                <CardTitle>Quote</CardTitle>
                <CardDescription>
                  {quote.from_amount} {quote.from_asset} → {quote.to_amount} {quote.to_asset}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div className="text-muted-foreground">Rate</div>
                  <div className="font-mono text-right">{quote.rate}</div>
                  <div className="text-muted-foreground">Fee</div>
                  <div className="font-mono text-right">{quote.fee}</div>
                  <div className="text-muted-foreground">Expires</div>
                  <div className="text-right text-xs">
                    {new Date(quote.expires_at).toLocaleString()}
                  </div>
                </div>
                {rate && (
                  <div className="rounded-lg border border-border bg-muted p-3 text-xs flex items-center justify-between">
                    <span>Provider {rate.provider}</span>
                    <Badge variant={rate.stale ? 'warning' : 'success'}>
                      {rate.spread_bps} bps
                    </Badge>
                  </div>
                )}
                <Button onClick={handleConvert} isLoading={isConverting} disabled={!walletId}>
                  Execute Conversion
                </Button>
              </CardContent>
            </Card>
          )}

          {!quote && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Coins className="h-5 w-5" />
                  Tip
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground">
                  Use FX page for live rates. Conversions are also available at{' '}
                  <span className="font-mono">/fx</span>.
                </p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
