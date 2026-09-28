'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import { api, type ScheduleResponse } from '@/lib/api';
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
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Calendar, Plus, Pause, Play, Trash2 } from 'lucide-react';

export default function SchedulesPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  const [schedules, setSchedules] = useState<ScheduleResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({
    from_wallet_id: '',
    to_wallet_id: '',
    asset: 'XLM',
    amount: '',
    frequency: 'weekly' as 'daily' | 'weekly' | 'monthly',
    start_date: new Date().toISOString().slice(0, 16),
    end_date: '',
  });

  const fetchSchedules = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.listSchedules();
      setSchedules(res.schedules || []);
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to load schedules', 'error');
    } finally {
      setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await fetchSchedules();
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchSchedules]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const startIso = new Date(form.start_date).toISOString();
      const endIso = form.end_date ? new Date(form.end_date).toISOString() : undefined;
      await api.createSchedule({
        from_wallet_id: form.from_wallet_id,
        to_wallet_id: form.to_wallet_id,
        asset: form.asset,
        amount: form.amount,
        frequency: form.frequency,
        start_date: startIso,
        end_date: endIso,
      });
      toast('Schedule created', 'success');
      setShowForm(false);
      await fetchSchedules();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Create failed', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggle = async (s: ScheduleResponse) => {
    try {
      const next = s.status === 'active' ? 'paused' : 'active';
      await api.updateSchedule(s.id, { status: next });
      toast(`Schedule ${next}`, 'success');
      await fetchSchedules();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Update failed', 'error');
    }
  };

  const handleCancel = async (id: string) => {
    if (!confirm('Cancel this schedule?')) return;
    try {
      await api.cancelSchedule(id);
      toast('Schedule cancelled', 'success');
      await fetchSchedules();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Cancel failed', 'error');
    }
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Scheduled Payouts"
        description="Recurring transfers — daily, weekly, monthly. Worker checks every minute."
      >
        <Button onClick={() => setShowForm(!showForm)} variant={showForm ? 'secondary' : 'primary'}>
          {showForm ? (
            'Cancel'
          ) : (
            <>
              <Plus className="h-4 w-4" /> New Schedule
            </>
          )}
        </Button>
      </PageHeader>

      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle>Create Schedule</CardTitle>
            <CardDescription>End date is optional — leave blank for indefinite.</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreate} className="flex flex-col gap-5">
              <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">From Wallet</label>
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
                  <label className="text-sm font-medium">To Wallet</label>
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
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Asset</label>
                  <Input
                    value={form.asset}
                    onChange={(e) => setForm({ ...form, asset: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Amount</label>
                  <Input
                    value={form.amount}
                    onChange={(e) => setForm({ ...form, amount: e.target.value })}
                    required
                    className="font-mono"
                  />
                </div>
              </div>
              <div className="grid grid-cols-1 gap-5 md:grid-cols-3">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Frequency</label>
                  <Select
                    value={form.frequency}
                    onChange={(e) =>
                      setForm({
                        ...form,
                        frequency: e.target.value as 'daily' | 'weekly' | 'monthly',
                      })
                    }
                  >
                    <option value="daily">Daily</option>
                    <option value="weekly">Weekly</option>
                    <option value="monthly">Monthly</option>
                  </Select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Start Date</label>
                  <Input
                    type="datetime-local"
                    value={form.start_date}
                    onChange={(e) => setForm({ ...form, start_date: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">End Date (optional)</label>
                  <Input
                    type="datetime-local"
                    value={form.end_date}
                    onChange={(e) => setForm({ ...form, end_date: e.target.value })}
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button type="submit" isLoading={submitting}>
                  Create
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <Card>
        {schedules.length === 0 ? (
          <EmptyState
            icon={Calendar}
            title="No schedules"
            description="Create a recurring payout to automate transfers."
          />
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableHeader>ID</TableHeader>
                <TableHeader>From → To</TableHeader>
                <TableHeader>Amount</TableHeader>
                <TableHeader>Frequency</TableHeader>
                <TableHeader>Next Run</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader className="text-right">Actions</TableHeader>
              </TableRow>
            </TableHead>
            <TableBody>
              {schedules.map((s) => (
                <TableRow key={s.id}>
                  <TableCell className="font-mono text-xs">{s.id.slice(0, 8)}</TableCell>
                  <TableCell className="font-mono text-xs">
                    {s.from_wallet_id.slice(0, 6)}... → {s.to_wallet_id.slice(0, 6)}...
                  </TableCell>
                  <TableCell>
                    {s.amount} {s.asset}
                  </TableCell>
                  <TableCell>{s.frequency}</TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {new Date(s.next_run_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    {s.status === 'active' ? (
                      <Badge variant="success">active</Badge>
                    ) : s.status === 'paused' ? (
                      <Badge variant="warning">paused</Badge>
                    ) : (
                      <Badge variant="default">{s.status}</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right flex justify-end gap-2">
                    {(s.status === 'active' || s.status === 'paused') && (
                      <Button variant="ghost" size="sm" onClick={() => handleToggle(s)}>
                        {s.status === 'active' ? (
                          <Pause className="h-3.5 w-3.5" />
                        ) : (
                          <Play className="h-3.5 w-3.5" />
                        )}
                        {s.status === 'active' ? 'Pause' : 'Resume'}
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-danger hover:text-danger"
                      onClick={() => handleCancel(s.id)}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Cancel
                    </Button>
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
