'use client';

import { useEffect, useState, useCallback } from 'react';
import { api, type APIKey } from '@/lib/api';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Modal } from '@/components/ui/modal';
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
import { KeyRound, Plus, Copy, Trash2 } from 'lucide-react';

export default function ApiKeysPage() {
  const { toast } = useToast();
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [label, setLabel] = useState('');
  const [mode, setMode] = useState<'live' | 'test'>(
    () => (typeof window !== 'undefined' && window.localStorage.getItem('fluxa_mode') === 'test' ? 'test' : 'live'),
  );

  const fetchKeys = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listAPIKeys();
      setKeys(Array.isArray(data) ? data : []);
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to load API keys', 'error');
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await fetchKeys();
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchKeys]);

  const handleCreateKey = async () => {
    setCreating(true);
    try {
      const res = await api.createAPIKey(label || undefined, mode);
      setNewKey(res.key);
      setLabel('');
      toast('API key created', 'success');
      await fetchKeys();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to create key', 'error');
    } finally {
      setCreating(false);
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this key?')) return;
    try {
      await api.revokeAPIKey(id);
      toast('API key revoked', 'success');
      await fetchKeys();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to revoke key', 'error');
    }
  };

  const copyKey = () => {
    if (newKey) {
      navigator.clipboard.writeText(newKey);
      toast('Key copied to clipboard', 'success');
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-40" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="API Keys"
        description="Manage your API keys for authenticating requests to Fluxa."
      >
        <div className="flex items-center gap-3">
          <Input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="Key label (optional)"
            className="w-48"
          />
          <Select value={mode} onChange={(e) => setMode(e.target.value as 'live' | 'test')} className="w-36" aria-label="API key environment">
            <option value="live">Live · mainnet</option>
            <option value="test">Test · testnet</option>
          </Select>
          <Button onClick={handleCreateKey} isLoading={creating}>
            <Plus className="h-4 w-4" />
            Create Secret Key
          </Button>
        </div>
      </PageHeader>

      <Modal
        open={!!newKey}
        onClose={() => setNewKey(null)}
        title="API Key Created"
        description="Please copy this key and save it somewhere safe. For security reasons, we cannot show it to you again."
      >
        <div className="flex flex-col gap-5">
          <div className="flex items-center justify-between gap-3 rounded-xl border border-primary/20 bg-primary-subtle p-4">
            <code className="break-all font-mono text-sm text-foreground">{newKey}</code>
            <Button variant="secondary" size="sm" onClick={copyKey}>
              <Copy className="h-4 w-4" />
              Copy
            </Button>
          </div>
          <Button onClick={() => setNewKey(null)} className="w-full">
            I have saved my key
          </Button>
        </div>
      </Modal>

      <Card>
        {keys.length === 0 ? (
          <EmptyState
            icon={KeyRound}
            title="No API keys yet"
            description="Create your first key to authenticate API requests."
          />
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableHeader>Name</TableHeader>
                <TableHeader>Environment</TableHeader>
                <TableHeader>Token</TableHeader>
                <TableHeader>Created</TableHeader>
                <TableHeader>Last Used</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader className="text-right">Actions</TableHeader>
              </TableRow>
            </TableHead>
            <TableBody>
              {keys.map((k) => {
                const isRevoked = !!k.revoked_at;
                return (
                  <TableRow
                    key={k.id}
                    className={isRevoked ? 'opacity-60' : undefined}
                  >
                    <TableCell className="font-medium">
                      {k.label || 'Unnamed Key'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={k.mode === 'test' ? 'warning' : 'default'}>{k.mode}</Badge>
                    </TableCell>
                    <TableCell>
                      <code className="rounded-md border border-border bg-muted px-2 py-1 font-mono text-xs">
                        {k.prefix}••••••••••••
                      </code>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(k.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {k.last_used_at
                        ? new Date(k.last_used_at).toLocaleDateString()
                        : 'Never'}
                    </TableCell>
                    <TableCell>
                      {isRevoked ? (
                        <Badge variant="default">Revoked</Badge>
                      ) : (
                        <Badge variant="success">Active</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      {!isRevoked && (
                        <Button
                          variant="ghost"
                          size="sm"
                          className="text-danger hover:bg-danger-subtle hover:text-danger"
                          onClick={() => handleRevoke(k.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                          Revoke
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </Card>
    </div>
  );
}
