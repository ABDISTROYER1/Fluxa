'use client';

import { FlaskConical, Radio } from 'lucide-react';
import { useAuth, type EnvironmentMode } from '@/lib/auth-context';

export function EnvironmentBanner() {
  const { mode, switchMode } = useAuth();

  const toggle = () => {
    const next: EnvironmentMode = mode === 'test' ? 'live' : 'test';
    switchMode(next);
    // Mode-scoped lists and credentials must be re-fetched after switching.
    window.location.reload();
  };

  if (mode !== 'test') {
    return (
      <div className="flex items-center justify-between border-b border-border bg-surface px-6 py-2 text-xs text-muted-foreground">
        <span className="flex items-center gap-2"><Radio className="h-3.5 w-3.5" /> Live environment · Stellar mainnet</span>
        <button type="button" onClick={toggle} className="font-semibold text-primary hover:underline">Switch to test mode</button>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-between bg-warning px-6 py-2 text-xs font-semibold text-foreground">
      <span className="flex items-center gap-2"><FlaskConical className="h-4 w-4" /> Test Mode · Stellar testnet · No real funds are used</span>
      <button type="button" onClick={toggle} className="rounded border border-foreground/20 px-2 py-1 hover:bg-background/20">Switch to live</button>
    </div>
  );
}
