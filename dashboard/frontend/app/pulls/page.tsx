'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { GitPullRequest, Search } from 'lucide-react';
import clsx from 'clsx';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { timeAgo, shortSha } from '@/lib/format';
import type { PullRecord } from '@/lib/types';
import { Badge, EmptyState, ErrorState, Panel, Skeleton } from '@/components/ui';/**
 * PR status is derived from the linked issue's pipeline stage:
 * - work-done   → agent self-review is queued/running (or a fix cycle) → "Being reviewed"
 * - review-done → self-review passed, posted "ready for human review and merge"
 * - failed      → max review iterations reached, posted "Please review manually"
 */
type PrStatus = 'reviewing' | 'human' | 'other';

const STATUS_META: Record<PrStatus, { label: string; color: string }> = {
  reviewing: { label: 'Being reviewed', color: '#38bdf8' },
  human: { label: 'Human acceptance', color: '#fbbf24' },
  other: { label: 'Other', color: '#94a3b8' },
};

function statusOf(stage: string | undefined): PrStatus {
  if (stage === 'work-done') return 'reviewing';
  if (stage === 'review-done' || stage === 'failed') return 'human';
  return 'other';
}

function PullCard({ pr, status, iterations }: { pr: PullRecord; status: PrStatus; iterations: number }) {
  const meta = STATUS_META[status];
  const openPr = (e: React.MouseEvent) => {
    e.preventDefault();
    if (pr.url) window.open(pr.url, '_blank', 'noopener');
  };
  return (
    <div onClick={openPr} className="panel panel-hover block cursor-pointer p-3.5">
      <div className="flex items-start justify-between gap-2">
        <a
          href={pr.url || '#'}
          onClick={(e) => e.stopPropagation()}
          target="_blank"
          rel="noreferrer"
          className="flex items-center gap-1.5 font-mono text-xs text-ink-faint transition hover:text-accent"
        >
          <GitPullRequest size={13} className="text-fuchsia" />
          #{pr.number}
        </a>
        <Badge color={meta.color} dot>
          {meta.label}
        </Badge>
      </div>
      <a
        href={pr.url || '#'}
        onClick={(e) => e.stopPropagation()}
        target="_blank"
        rel="noreferrer"
        className="mt-2 line-clamp-2 block text-sm font-medium text-ink transition hover:text-accent"
      >
        {pr.title}
      </a>
      <div className="mt-3 flex items-center justify-between text-[0.68rem] text-ink-faint">
        <Link
          href={`/issues/${pr.issue_number}`}
          onClick={(e) => e.stopPropagation()}
          className="font-mono text-accent hover:underline"
        >
          issue #{pr.issue_number}
        </Link>
        <span className="font-mono">{shortSha(pr.head_sha)}</span>
      </div>
      <div className="mt-1 flex items-center justify-between">
        <p className="text-[0.65rem] text-ink-faint">updated {timeAgo(pr.updated_at)}</p>
        {iterations > 0 && (
          <p className="text-[0.65rem] text-ink-faint">
            {iterations} review round{iterations > 1 ? 's' : ''}
          </p>
        )}
      </div>
    </div>
  );
}

const FILTERS: Array<{ key: PrStatus | 'all'; label: string }> = [
  { key: 'all', label: 'All' },
  { key: 'reviewing', label: 'Being reviewed' },
  { key: 'human', label: 'Human acceptance' },
];

export default function PullsPage() {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<PrStatus | 'all'>('all');
  const { data: pullsData, error, refresh } = useInstanceData((base) => api.pulls(base, 500), 15000);
  const { data: issuesData } = useInstanceData((base) => api.issues(base, { limit: 500 }), 30000);

  const issueById = useMemo(() => {
    const m = new Map<number, { stage: string; review_iterations: number }>();
    for (const i of issuesData?.issues ?? []) m.set(i.number, { stage: i.stage, review_iterations: i.review_iterations });
    return m;
  }, [issuesData]);

  const openPulls = useMemo(
    () => (pullsData?.pulls ?? []).filter((p) => p.state === 'open'),
    [pullsData],
  );

  const pulls = useMemo(() => {
    const q = query.trim().toLowerCase();
    return openPulls.filter((p) => {
      const status = statusOf(issueById.get(p.issue_number)?.stage);
      if (filter !== 'all' && status !== filter) return false;
      if (!q) return true;
      return (
        p.title.toLowerCase().includes(q) ||
        String(p.number).includes(q) ||
        String(p.issue_number).includes(q)
      );
    });
  }, [openPulls, issueById, query, filter]);

  const counts = useMemo(() => {
    const c: Record<string, number> = { all: openPulls.length, reviewing: 0, human: 0 };
    for (const p of openPulls) {
      const status = statusOf(issueById.get(p.issue_number)?.stage);
      c[status]++;
    }
    return c;
  }, [openPulls, issueById]);

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-ink">Pull Requests</h1>
          <p className="mt-1 text-xs text-ink-faint">{openPulls.length} open · awaiting merge</p>
        </div>
        <div className="relative w-full max-w-xs">
          <Search
            size={14}
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint"
          />
          <input
            className="input !pl-9"
            placeholder="Search open PRs…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      <div className="flex flex-wrap gap-1.5">
        {FILTERS.map((f) => (
          <button
            key={f.key}
            onClick={() => setFilter(f.key)}
            className={clsx(
              'chip transition',
              filter === f.key
                ? '!border-accent/50 !bg-accent/10 !text-accent'
                : 'hover:border-edge-2',
            )}
          >
            {f.label}
            <span className="text-ink-faint">{counts[f.key] ?? 0}</span>
          </button>
        ))}
      </div>

      {error && !pullsData ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !pullsData ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      ) : pulls.length === 0 ? (
        <Panel>
          <EmptyState
            title={openPulls.length ? 'No matching PRs' : 'No open PRs yet'}
            hint="PRs from implemented issues appear here until they are merged."
          />
        </Panel>
      ) : (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {pulls.map((pr) => {
            const issue = issueById.get(pr.issue_number);
            return (
              <PullCard
                key={pr.number}
                pr={pr}
                status={statusOf(issue?.stage)}
                iterations={issue?.review_iterations ?? 0}
              />
            );
          })}
        </div>
      )}
    </div>
  );
}