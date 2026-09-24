'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import clsx from 'clsx';
import { Filter } from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { ErrorState, Panel } from '@/components/ui';
import { PageHeader } from '@/components/page-header';
import { RunsView } from '@/components/runs-view';

const STATUS_FILTERS = ['all', 'running', 'completed', 'failed', 'interrupted'] as const;
type StatusFilter = (typeof STATUS_FILTERS)[number];

const STATUS_LABELS: Record<StatusFilter, string> = {
  all: 'All',
  running: 'Running',
  completed: 'Completed',
  failed: 'Failed',
  interrupted: 'Interrupted',
};

const STAGES = ['all', 'spec', 'work', 'review', 'fix', 'merge'];

export default function RunsPage() {
  return (
    <Suspense
      fallback={
        <div className="space-y-5">
          <div className="skeleton h-10 w-64" />
          <div className="skeleton h-96 w-full" />
        </div>
      }
    >
      <RunsPageInner />
    </Suspense>
  );
}

function RunsPageInner() {
  const searchParams = useSearchParams();
  const statusParam = searchParams.get('status') ?? 'all';
  const initialStatus = STATUS_FILTERS.includes(statusParam as StatusFilter)
    ? (statusParam as StatusFilter)
    : 'all';
  const [status, setStatus] = useState<StatusFilter>(initialStatus);
  const [stage, setStage] = useState('all');
  const { data, error, refresh } = useInstanceData((base) => api.runs(base, 1000), 15000);

  useEffect(() => {
    setStatus(
      STATUS_FILTERS.includes(statusParam as StatusFilter)
        ? (statusParam as StatusFilter)
        : 'all',
    );
  }, [statusParam]);

  const runs = data?.runs ?? [];
  const filtered = useMemo(
    () =>
      runs.filter(
        (run) =>
          (status === 'all' || run.status === status) &&
          (stage === 'all' || run.stage === stage),
      ),
    [runs, status, stage],
  );

  const statusCounts = useMemo(() => {
    const counts: Record<StatusFilter, number> = {
      all: runs.length,
      running: 0,
      completed: 0,
      failed: 0,
      interrupted: 0,
    };
    for (const run of runs) {
      if (run.status in counts) counts[run.status as StatusFilter] += 1;
    }
    return counts;
  }, [runs]);

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Runs"
        title="Run history"
        subtitle={`${filtered.length} shown · ${statusCounts.running} running · ${statusCounts.failed} failed`}
      />

      <div className="space-y-3 rounded-xl border border-edge bg-panel p-3">
        <div className="flex flex-wrap items-center gap-2">
          <Filter size={13} className="mr-1 text-ink-faint" />
          <span className="mr-1 text-[10px] font-semibold uppercase tracking-[0.1em] text-ink-faint">
            Status
          </span>
          {STATUS_FILTERS.map((item) => (
            <button
              key={item}
              type="button"
              onClick={() => setStatus(item)}
              className={clsx(
                'chip transition',
                status === item
                  ? '!border-accent/50 !bg-accent/10 !text-accent'
                  : 'text-ink-dim hover:border-edge-2 hover:text-ink',
              )}
              aria-pressed={status === item}
            >
              {STATUS_LABELS[item]}
              <span className="text-ink-faint">{statusCounts[item]}</span>
            </button>
          ))}
        </div>
        <div className="flex flex-wrap items-center gap-2 border-t border-edge pt-3">
          <span className="mr-1 text-[10px] font-semibold uppercase tracking-[0.1em] text-ink-faint">
            Stage
          </span>
          {STAGES.map((item) => (
            <button
              key={item}
              type="button"
              onClick={() => setStage(item)}
              className={clsx(
                'chip transition',
                stage === item
                  ? '!border-accent/50 !bg-accent/10 !text-accent'
                  : 'text-ink-dim hover:border-edge-2 hover:text-ink',
              )}
              aria-pressed={stage === item}
            >
              {item}
            </button>
          ))}
        </div>
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <Panel>
          <div className="space-y-3">
            {Array.from({ length: 8 }).map((_, index) => (
              <div key={index} className="skeleton h-12 w-full" />
            ))}
          </div>
        </Panel>
      ) : (
        <RunsView runs={filtered} />
      )}
    </div>
  );
}
