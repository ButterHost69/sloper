'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import { ExternalLink, GitMerge, GitPullRequest, Search } from 'lucide-react';
import clsx from 'clsx';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { shortSha, timeAgo } from '@/lib/format';
import type { PullRecord } from '@/lib/types';
import { PageHeader } from '@/components/page-header';
import { Badge, EmptyState, ErrorState, Panel, Skeleton } from '@/components/ui';

type WorkflowStatus = 'reviewing' | 'human' | 'other';
type PullStateFilter = 'all' | 'open' | 'merged' | 'closed';

const WORKFLOW_META: Record<WorkflowStatus, { label: string; color: string }> = {
  reviewing: { label: 'Being reviewed', color: '#60a5fa' },
  human: { label: 'Human acceptance', color: '#f5ad31' },
  other: { label: 'Other', color: '#94a3b8' },
};

const STATE_META: Record<Exclude<PullStateFilter, 'all'>, { label: string; color: string }> = {
  open: { label: 'Open', color: '#4ed17e' },
  merged: { label: 'Merged', color: '#a78bfa' },
  closed: { label: 'Closed', color: '#94a3b8' },
};

const STATE_FILTERS: Array<{ key: PullStateFilter; label: string }> = [
  { key: 'all', label: 'All states' },
  { key: 'open', label: 'Open' },
  { key: 'merged', label: 'Merged' },
  { key: 'closed', label: 'Closed' },
];

function workflowStatus(stage: string | undefined): WorkflowStatus {
  if (stage === 'work-done') return 'reviewing';
  if (stage === 'review-done' || stage === 'failed') return 'human';
  return 'other';
}

function PullCard({
  pr,
  workflow,
  iterations,
}: {
  pr: PullRecord;
  workflow: WorkflowStatus;
  iterations: number;
}) {
  const state = (pr.state === 'merged' || pr.state === 'closed' ? pr.state : 'open') as Exclude<
    PullStateFilter,
    'all'
  >;
  const stateMeta = STATE_META[state];
  const workflowMeta = WORKFLOW_META[workflow];

  return (
    <article className="panel panel-hover flex min-h-44 flex-col p-4">
      <div className="flex items-start justify-between gap-3">
        <a
          href={pr.url || '#'}
          target="_blank"
          rel="noreferrer"
          className="flex items-center gap-1.5 font-mono text-xs text-ink-faint transition hover:text-accent"
        >
          <GitPullRequest size={13} className="text-fuchsia" /> #{pr.number}
          <ExternalLink size={10} />
        </a>
        <Badge color={stateMeta.color} dot>
          {stateMeta.label}
        </Badge>
      </div>

      <a
        href={pr.url || '#'}
        target="_blank"
        rel="noreferrer"
        className="mt-3 line-clamp-2 text-sm font-medium leading-5 text-ink transition hover:text-accent"
      >
        {pr.title}
      </a>

      {state === 'open' && (
        <div className="mt-3">
          <Badge color={workflowMeta.color} dot>
            {workflowMeta.label}
          </Badge>
        </div>
      )}

      <div className="mt-auto flex items-end justify-between gap-3 border-t border-edge pt-3 text-[10px] text-ink-faint">
        <div className="min-w-0 space-y-1">
          <p>
            Updated {timeAgo(pr.updated_at)}
            {iterations > 0 ? ` · ${iterations} review round${iterations > 1 ? 's' : ''}` : ''}
          </p>
          <p className="flex items-center gap-2">
            <Link
              href={`/issues/${pr.issue_number}`}
              className="font-mono text-accent hover:underline"
            >
              issue #{pr.issue_number}
            </Link>
            <span className="font-mono">{shortSha(pr.head_sha)}</span>
          </p>
        </div>
        {state === 'merged' && <GitMerge size={14} className="shrink-0 text-violet" />}
      </div>
    </article>
  );
}

export default function PullsPage() {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<PullStateFilter>('all');
  const { data, error, refresh } = useInstanceData((base) => api.pulls(base, 500), 15000);
  const { data: issuesData } = useInstanceData((base) => api.issues(base, { limit: 500 }), 30000);

  const issueById = useMemo(() => {
    const map = new Map<number, { stage: string; review_iterations: number }>();
    for (const issue of issuesData?.issues ?? []) {
      map.set(issue.number, {
        stage: issue.stage,
        review_iterations: issue.review_iterations,
      });
    }
    return map;
  }, [issuesData]);

  const allPulls = data?.pulls ?? [];
  const counts = useMemo(
    () => ({
      all: allPulls.length,
      open: allPulls.filter((pull) => pull.state === 'open').length,
      merged: allPulls.filter((pull) => pull.state === 'merged').length,
      closed: allPulls.filter((pull) => pull.state === 'closed').length,
    }),
    [allPulls],
  );

  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return allPulls.filter((pull) => {
      if (filter !== 'all' && pull.state !== filter) return false;
      if (!normalized) return true;
      return [pull.title, String(pull.number), String(pull.issue_number)]
        .join(' ')
        .toLowerCase()
        .includes(normalized);
    });
  }, [allPulls, filter, query]);

  const groups = useMemo(
    () =>
      (['open', 'merged', 'closed'] as const)
        .map((state) => ({ state, pulls: filtered.filter((pull) => pull.state === state) }))
        .filter((group) => group.pulls.length > 0),
    [filtered],
  );

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Pull Requests"
        title="Pull requests"
        subtitle={`${counts.all} total · ${counts.open} open · ${counts.merged} merged`}
        actions={
          <div className="relative w-full lg:w-80">
            <Search
              size={14}
              className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint"
            />
            <input
              className="input !pl-9"
              aria-label="Search pull requests"
              placeholder="Search title, PR, or issue…"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
          </div>
        }
      />

      <div className="flex flex-wrap gap-1.5" aria-label="Pull request state filter">
        {STATE_FILTERS.map((item) => (
          <button
            key={item.key}
            type="button"
            onClick={() => setFilter(item.key)}
            className={clsx(
              'chip transition',
              filter === item.key
                ? '!border-accent/50 !bg-accent/10 !text-accent'
                : 'text-ink-dim hover:border-edge-2 hover:text-ink',
            )}
            aria-pressed={filter === item.key}
          >
            {item.label}
            <span className="text-ink-faint">{counts[item.key]}</span>
          </button>
        ))}
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton key={index} className="h-48 w-full" />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <Panel>
          <EmptyState
            icon={<GitPullRequest size={19} />}
            title={allPulls.length ? 'No matching pull requests' : 'No pull requests yet'}
            hint="Pull requests created by implemented issues appear here with their current GitHub state."
          />
        </Panel>
      ) : (
        <div className="space-y-7">
          {groups.map((group) => {
            const meta = STATE_META[group.state];
            return (
              <section key={group.state} aria-labelledby={`pulls-${group.state}`}>
                <div className="mb-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span
                      className="h-2 w-2 rounded-full"
                      style={{ backgroundColor: meta.color }}
                      aria-hidden="true"
                    />
                    <h2 id={`pulls-${group.state}`} className="text-sm font-semibold text-ink">
                      {meta.label}
                    </h2>
                    <span className="text-[10px] text-ink-faint">{group.pulls.length}</span>
                  </div>
                </div>
                <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 2xl:grid-cols-3">
                  {group.pulls.map((pull) => {
                    const issue = issueById.get(pull.issue_number);
                    return (
                      <PullCard
                        key={pull.number}
                        pr={pull}
                        workflow={workflowStatus(issue?.stage)}
                        iterations={issue?.review_iterations ?? 0}
                      />
                    );
                  })}
                </div>
              </section>
            );
          })}
        </div>
      )}
    </div>
  );
}
