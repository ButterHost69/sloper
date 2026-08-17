'use client';

import { Suspense, useMemo, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { ArrowUpRight, GitPullRequest, Search } from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { timeAgo, formatTime } from '@/lib/format';
import { STAGES, stageMeta } from '@/lib/stages';
import type { Issue } from '@/lib/types';
import { EmptyState, ErrorState, Panel, Skeleton } from '@/components/ui';
import { LabelChips, StageBadge } from '@/components/badges';
import clsx from 'clsx';

function IssueRow({ issue }: { issue: Issue }) {
  const meta = stageMeta(issue.stage);
  return (
    <Link
      href={`/issues/${issue.number}`}
      className="group grid grid-cols-12 items-center gap-3 border-b border-edge/60 px-4 py-3 transition hover:bg-white/[0.03]"
    >
      <div className="col-span-12 flex items-center gap-2 sm:col-span-6">
        <span
          className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border font-mono text-xs font-semibold"
          style={{
            color: meta.color,
            borderColor: `${meta.color}33`,
            background: `${meta.color}12`,
          }}
        >
          {issue.number}
        </span>
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-ink group-hover:text-accent">
            {issue.title}
          </p>
          <div className="mt-0.5 flex items-center gap-2 text-[0.68rem] text-ink-faint">
            <span>@{issue.author}</span>
            <span>·</span>
            <span>{timeAgo(issue.updated_at_local)}</span>
            {issue.pr_number > 0 && (
              <span className="flex items-center gap-0.5 text-fuchsia">
                <GitPullRequest size={10} /> #{issue.pr_number}
              </span>
            )}
          </div>
        </div>
      </div>
      <div className="col-span-4 sm:col-span-2">
        <StageBadge stage={issue.stage} />
      </div>
      <div className="col-span-4 sm:col-span-2">
        <LabelChips labels={issue.labels} />
      </div>
      <div className="col-span-4 flex items-center justify-end gap-2 sm:col-span-2">
        <span className="hidden text-[0.68rem] text-ink-faint md:block">
          {formatTime(issue.created_at)}
        </span>
        <ArrowUpRight size={14} className="text-ink-faint opacity-0 transition group-hover:opacity-100" />
      </div>
    </Link>
  );
}

export default function IssuesPage() {
  return (
    <Suspense fallback={<div className="space-y-5"><div className="skeleton h-10 w-64" /><div className="skeleton h-96 w-full" /></div>}>
      <IssuesPageInner />
    </Suspense>
  );
}

function IssuesPageInner() {
  const searchParams = useSearchParams();
  const initialStage = searchParams.get('stage') ?? 'all';
  const [query, setQuery] = useState('');
  const [stage, setStage] = useState(initialStage);

  const { data, error, refresh } = useInstanceData(
    (base) => api.issues(base, 1000),
    15000,
  );

  const filtered = useMemo(() => {
    const issues = data?.issues ?? [];
    const q = query.trim().toLowerCase();
    return issues.filter((i) => {
      if (stage !== 'all' && i.stage !== stage) return false;
      if (!q) return true;
      return (
        i.title.toLowerCase().includes(q) ||
        String(i.number).includes(q) ||
        i.author.toLowerCase().includes(q) ||
        i.labels.some((l) => l.toLowerCase().includes(q))
      );
    });
  }, [data, query, stage]);

  const counts = useMemo(() => {
    const byStage: Record<string, number> = {};
    for (const i of data?.issues ?? []) byStage[i.stage] = (byStage[i.stage] ?? 0) + 1;
    return byStage;
  }, [data]);

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-ink">Issues</h1>
          <p className="mt-1 text-xs text-ink-faint">
            {data?.count ?? 0} cached issues · {filtered.length} shown
          </p>
        </div>
        <div className="relative w-full max-w-xs">
          <Search size={14} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint" />
          <input
            className="input !pl-9"
            placeholder="Search title, #number, author, label…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      {/* Stage filter chips */}
      <div className="flex flex-wrap gap-1.5">
        <button
          onClick={() => setStage('all')}
          className={clsx(
            'chip transition',
            stage === 'all' ? '!border-accent/50 !bg-accent/10 !text-accent' : 'hover:border-edge-2',
          )}
        >
          All
          <span className="text-ink-faint">{data?.count ?? 0}</span>
        </button>
        {STAGES.map((s) => (
          <button
            key={s.key}
            onClick={() => setStage(s.key)}
            className={clsx('chip transition', stage === s.key && '!border-accent/50 !bg-accent/10 !text-accent')}
            style={stage === s.key ? undefined : { color: s.color }}
          >
            {s.label}
            <span className="text-ink-faint">{counts[s.key] ?? 0}</span>
          </button>
        ))}
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <Panel className="p-0">
          <div className="space-y-3 p-4">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))}
          </div>
        </Panel>
      ) : (
        <Panel className="overflow-hidden p-0">
          {filtered.length === 0 ? (
            <EmptyState title="No matching issues" hint="Adjust the search or filter." />
          ) : (
            <div>
              <div className="grid grid-cols-12 gap-3 border-b border-edge bg-panel-2 px-4 py-2 text-[0.65rem] font-semibold uppercase tracking-wider text-ink-faint">
                <span className="col-span-6">Issue</span>
                <span className="col-span-2">Stage</span>
                <span className="col-span-2">Labels</span>
                <span className="col-span-2 text-right">Created</span>
              </div>
              {filtered.map((i) => (
                <IssueRow key={i.number} issue={i} />
              ))}
            </div>
          )}
        </Panel>
      )}
    </div>
  );
}