'use client';

import { Suspense, useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import clsx from 'clsx';
import { Filter } from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { EmptyState, ErrorState, Panel, StaleDataNotice } from '@/components/ui';
import { useInstances } from '@/components/instance-context';
import { PageHeader } from '@/components/page-header';
import { RunsView } from '@/components/runs-view';
import type { RunRecord } from '@/lib/types';

const STATUS_FILTERS = ['all', 'attention', 'running', 'completed', 'failed', 'interrupted'] as const;
type StatusFilter = (typeof STATUS_FILTERS)[number];

const STATUS_LABELS: Record<StatusFilter, string> = {
  all: 'All',
  attention: 'Needs attention',
  running: 'Running',
  completed: 'Completed',
  failed: 'Failed',
  interrupted: 'Interrupted',
};

const STAGES = ['all', 'spec', 'work', 'review', 'fix', 'merge'];
const PAGE_SIZE = 500;

const isAttentionStatus = (status: string) => status === 'failed' || status === 'interrupted';

function matchesStatus(runStatus: string, filter: StatusFilter): boolean {
  if (filter === 'all') return true;
  if (filter === 'attention') return isAttentionStatus(runStatus);
  return runStatus === filter;
}

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
  const [olderRuns, setOlderRuns] = useState<RunRecord[]>([]);
  const [loadingMore, setLoadingMore] = useState(false);
  const [moreError, setMoreError] = useState<Error | null>(null);
  const [exhausted, setExhausted] = useState(false);
  const { active } = useInstances();
  const activeUrlRef = useRef(active?.url);
  const { data, error, refresh } = useInstanceData(
    (base) => api.runs(base, PAGE_SIZE),
    15000,
  );

  useEffect(() => {
    setStatus(
      STATUS_FILTERS.includes(statusParam as StatusFilter)
        ? (statusParam as StatusFilter)
        : 'all',
    );
  }, [statusParam]);

  useEffect(() => {
    activeUrlRef.current = active?.url;
    setOlderRuns([]);
    setMoreError(null);
    setExhausted(false);
    setLoadingMore(false);
  }, [active?.url]);

  const hasMore = !exhausted && (data?.runs.length ?? 0) === PAGE_SIZE;

  const runs = useMemo(() => {
    const byId = new Map<number, RunRecord>();
    for (const run of [...(data?.runs ?? []), ...olderRuns]) byId.set(run.id, run);
    return [...byId.values()];
  }, [data, olderRuns]);

  const loadMore = async () => {
    if (!active || loadingMore || !hasMore) return;
    const requestUrl = active.url;
    setLoadingMore(true);
    setMoreError(null);
    try {
      const cursor = runs[runs.length - 1]?.id;
      if (!cursor) return;
      const next = await api.runs(requestUrl, PAGE_SIZE, 0, cursor);
      if (activeUrlRef.current !== requestUrl) return;
      setOlderRuns((current) => [...current, ...next.runs]);
      setExhausted(next.runs.length < PAGE_SIZE);
    } catch (loadError) {
      if (activeUrlRef.current === requestUrl) {
        setMoreError(loadError instanceof Error ? loadError : new Error('Could not load older runs'));
      }
    } finally {
      if (activeUrlRef.current === requestUrl) setLoadingMore(false);
    }
  };

  const filtered = useMemo(
    () =>
      runs.filter(
        (run) =>
          matchesStatus(run.status, status) &&
          (stage === 'all' || run.stage === stage),
      ),
    [runs, status, stage],
  );

  const hasFilters = status !== 'all' || stage !== 'all';
  const statusCounts = useMemo(() => {
    const counts: Record<StatusFilter, number> = {
      all: runs.length,
      attention: 0,
      running: 0,
      completed: 0,
      failed: 0,
      interrupted: 0,
    };
    for (const run of runs) {
      if (isAttentionStatus(run.status)) counts.attention += 1;
      if (run.status in counts) counts[run.status as StatusFilter] += 1;
    }
    return counts;
  }, [runs]);

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Runs"
        title="Run history"
        subtitle={`${filtered.length} shown · ${runs.length} loaded · ${statusCounts.running} running · ${statusCounts.failed} failed`}
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

      {error && data && <StaleDataNotice error={error} onRetry={refresh} label="run" />}

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        active ? (
          <Panel>
            <div className="space-y-3">
              {Array.from({ length: 8 }).map((_, index) => (
                <div key={index} className="skeleton h-12 w-full" />
              ))}
            </div>
          </Panel>
        ) : (
          <Panel bodyClassName="p-0">
            <EmptyState
              title="Connect an instance to inspect runs"
              hint="Run history and captured agent output come from the read-only Sloper API."
              action={
                <Link href="/instances" className="btn btn-primary mt-1">
                  Configure an instance
                </Link>
              }
            />
          </Panel>
        )
      ) : (
        <>
          {filtered.length === 0 ? (
            <Panel bodyClassName="p-0">
              <EmptyState
                title={runs.length ? 'No runs match these filters' : 'No runs yet'}
                hint={
                  runs.length
                    ? 'Try another status or stage filter.'
                    : 'Pipeline executions will show up here.'
                }
                action={
                  hasFilters ? (
                    <button
                      type="button"
                      className="btn mt-1"
                      onClick={() => {
                        setStatus('all');
                        setStage('all');
                      }}
                    >
                      Clear filters
                    </button>
                  ) : (
                    <button type="button" className="btn mt-1" onClick={() => void refresh()}>
                      Refresh instance
                    </button>
                  )
                }
              />
            </Panel>
          ) : (
            <RunsView runs={filtered} />
          )}
          {(hasMore || loadingMore || moreError) && (
            <div className="mt-3 flex flex-col items-center gap-2">
              {moreError && <p className="text-xs text-danger">{moreError.message}</p>}
              <button
                type="button"
                className="btn"
                onClick={() => void loadMore()}
                disabled={loadingMore || !hasMore}
              >
                {loadingMore ? 'Loading older runs…' : 'Load older runs'}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
