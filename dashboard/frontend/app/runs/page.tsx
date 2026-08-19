'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import clsx from 'clsx';
import { Filter } from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { ErrorState, Panel } from '@/components/ui';
import { RunsView } from '@/components/runs-view';
import type { RunRecord } from '@/lib/types';

const STATUSES = ['all', 'running', 'completed', 'failed', 'interrupted'];
const STAGES = ['all', 'spec', 'work', 'review', 'fix', 'merge'];

export default function RunsPage() {
  return (
    <Suspense fallback={<div className="space-y-5"><div className="skeleton h-10 w-64" /><div className="skeleton h-96 w-full" /></div>}>
      <RunsPageInner />
    </Suspense>
  );
}

function RunsPageInner() {
  const searchParams = useSearchParams();
  const statusParam = searchParams.get('status') ?? 'all';
  const [status, setStatus] = useState(statusParam);
  const [stage, setStage] = useState('all');
  const { data, error, refresh } = useInstanceData((base) => api.runs(base, 1000), 15000);

  // Keep the status filter in sync when the URL query param changes.
  useEffect(() => {
    setStatus(statusParam);
  }, [statusParam]);

  const filtered = useMemo(() => {
    const runs = data?.runs ?? [];
    return runs.filter(
      (r) =>
        (status === 'all' || r.status === status) &&
        (stage === 'all' || r.stage === stage),
    );
  }, [data, status, stage]);

  const statusCounts = useMemo(() => {
    const c: Record<string, number> = {};
    for (const r of data?.runs ?? []) c[r.status] = (c[r.status] ?? 0) + 1;
    return c;
  }, [data]);

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-ink">Runs</h1>
        <p className="mt-1 text-xs text-ink-faint">
          {data?.count ?? 0} pipeline executions · {filtered.length} shown
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
        <div className="flex items-center gap-1.5">
          <Filter size={13} className="text-ink-faint" />
          <span className="text-xs font-semibold uppercase tracking-wider text-ink-faint">Status</span>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {STATUSES.map((s) => (
            <button
              key={s}
              onClick={() => setStatus(s)}
              className={clsx(
                'chip transition',
                status === s ? '!border-accent/50 !bg-accent/10 !text-accent' : 'hover:border-edge-2',
              )}
            >
              {s}
              {s !== 'all' && <span className="text-ink-faint">{statusCounts[s] ?? 0}</span>}
            </button>
          ))}
        </div>
        <div className="ml-auto flex flex-wrap gap-1.5">
          {STAGES.map((s) => (
            <button
              key={s}
              onClick={() => setStage(s)}
              className={clsx(
                'chip transition',
                stage === s ? '!border-accent/50 !bg-accent/10 !text-accent' : 'hover:border-edge-2',
              )}
            >
              {s}
            </button>
          ))}
        </div>
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <Panel className="p-0">
          <div className="space-y-3 p-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="skeleton h-12 w-full" />
            ))}
          </div>
        </Panel>
      ) : (
        <RunsView runs={filtered as RunRecord[]} defaultOpenFirst />
      )}
    </div>
  );
}