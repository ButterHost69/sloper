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

const STATUS_FILTERS = ['all', 'ongoing', 'failed'] as const;
type StatusFilter = (typeof STATUS_FILTERS)[number];
const STAGES = ['all', 'spec', 'work', 'review', 'fix', 'merge'];

// The runs page is a monitor: show only what's in flux.
// ongoing = actually running; failed + interrupted = dead runs needing attention
// (interrupted = killed when sloper stopped, recovered on next boot).
const isOngoing = (status: string) => status === 'running';
const isFailed = (status: string) => status === 'failed' || status === 'interrupted';
const keepsList = (status: string) => isOngoing(status) || isFailed(status);

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
  const [status, setStatus] = useState<StatusFilter>(
    STATUS_FILTERS.includes(statusParam as StatusFilter) ? (statusParam as StatusFilter) : 'all',
  );
  const [stage, setStage] = useState('all');
  const { data, error, refresh } = useInstanceData((base) => api.runs(base, 1000), 15000);

  // Keep the status filter in sync when the URL query param changes.
  useEffect(() => {
    setStatus(
      STATUS_FILTERS.includes(statusParam as StatusFilter) ? (statusParam as StatusFilter) : 'all',
    );
  }, [statusParam]);

  const filtered = useMemo(() => {
    const runs = data?.runs ?? [];
    return runs.filter(
      (r) =>
        (status === 'all' ? keepsList(r.status) : status === 'ongoing' ? isOngoing(r.status) : isFailed(r.status)) &&
        (stage === 'all' || r.stage === stage),
    );
  }, [data, status, stage]);

  const statusCounts = useMemo(() => {
    let ongoing = 0;
    let failed = 0;
    for (const r of data?.runs ?? []) {
      if (isOngoing(r.status)) ongoing++;
      else if (isFailed(r.status)) failed++;
    }
    return { ongoing, failed, all: ongoing + failed };
  }, [data]);

  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-ink">Runs</h1>
        <p className="mt-1 text-xs text-ink-faint">
          {filtered.length} in flight · completed runs hidden
        </p>
      </div>

      <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
        <div className="flex items-center gap-1.5">
          <Filter size={13} className="text-ink-faint" />
          <span className="text-xs font-semibold uppercase tracking-wider text-ink-faint">Status</span>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {STATUS_FILTERS.map((s) => (
            <button
              key={s}
              onClick={() => setStatus(s)}
              className={clsx(
                'chip transition',
                status === s ? '!border-accent/50 !bg-accent/10 !text-accent' : 'hover:border-edge-2',
              )}
            >
              {s}
              <span className="text-ink-faint">{statusCounts[s] ?? 0}</span>
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
        <RunsView runs={filtered as RunRecord[]} />
      )}
    </div>
  );
}