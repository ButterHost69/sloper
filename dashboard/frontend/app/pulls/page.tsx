'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { GitPullRequest, Search } from 'lucide-react';
import clsx from 'clsx';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { timeAgo, shortSha } from '@/lib/format';
import type { PullRecord } from '@/lib/types';
import { EmptyState, ErrorState, Panel, Skeleton } from '@/components/ui';
import { ReviewStateBadge } from '@/components/badges';

const BOARD_COLUMNS = [
  { key: 'open', label: 'Open', color: '#38bdf8' },
  { key: 'merged', label: 'Merged', color: '#34d399' },
  { key: 'closed', label: 'Closed', color: '#71717a' },
];

function PullCard({ pr }: { pr: PullRecord }) {
  return (
    <Link
      href={pr.url || '#'}
      target={pr.url ? '_blank' : undefined}
      rel="noreferrer"
      className="panel panel-hover block p-3.5"
    >
      <div className="flex items-start justify-between gap-2">
        <span className="flex items-center gap-1.5 font-mono text-xs text-ink-faint">
          <GitPullRequest size={13} className="text-fuchsia" />
          #{pr.number}
        </span>
        <ReviewStateBadge state={pr.review_state} />
      </div>
      <p className="mt-2 line-clamp-2 text-sm font-medium text-ink">{pr.title}</p>
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
      <p className="mt-1 text-[0.65rem] text-ink-faint">{timeAgo(pr.updated_at)}</p>
    </Link>
  );
}

export default function PullsPage() {
  const [query, setQuery] = useState('');
  const { data, error, refresh } = useInstanceData((base) => api.pulls(base, 500), 15000);

  const filtered = useMemo(() => {
    const pulls = data?.pulls ?? [];
    const q = query.trim().toLowerCase();
    if (!q) return pulls;
    return pulls.filter(
      (p) =>
        p.title.toLowerCase().includes(q) ||
        String(p.number).includes(q) ||
        String(p.issue_number).includes(q),
    );
  }, [data, query]);

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-ink">Pull Requests</h1>
          <p className="mt-1 text-xs text-ink-faint">
            {data?.count ?? 0} tracked PRs
          </p>
        </div>
        <div className="relative w-full max-w-xs">
          <Search size={14} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint" />
          <input
            className="input !pl-9"
            placeholder="Search PRs…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {BOARD_COLUMNS.map((col) => {
            const items = filtered.filter((p) => p.state === col.key);
            return (
              <div key={col.key}>
                <div className="mb-2 flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full" style={{ background: col.color }} />
                  <h2 className="text-xs font-semibold uppercase tracking-wider text-ink-dim">
                    {col.label}
                  </h2>
                  <span className="chip !text-[0.62rem] text-ink-faint">{items.length}</span>
                </div>
                <div className="space-y-2.5">
                  {items.length === 0 ? (
                    <Panel>
                      <p className="py-6 text-center text-xs text-ink-faint">No {col.label.toLowerCase()} PRs</p>
                    </Panel>
                  ) : (
                    items.map((pr) => <PullCard key={pr.number} pr={pr} />)
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {filtered.length === 0 && !data && !error && (
        <Panel>
          <EmptyState title="No matching pull requests" />
        </Panel>
      )}
    </div>
  );
}