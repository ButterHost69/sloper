'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { GitBranch } from 'lucide-react';
import clsx from 'clsx';
import { useInstances } from '@/components/instance-context';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { stageMeta } from '@/lib/stages';
import { PulseDot } from '@/components/ui';
import type { Worktree, WorktreeIssue, WorktreeKind, WorktreesResponse } from '@/lib/types';

const KIND_META: Record<WorktreeKind, { label: string; color: string }> = {
  work: { label: 'work', color: '#a78bfa' },
  review: { label: 'review', color: '#e879f9' },
  fix: { label: 'fix', color: '#fbbf24' },
  unknown: { label: 'worktree', color: '#94a3b8' },
};

/** Fresh while an agent is working, cheap when the instance is idle. */
const BUSY_INTERVAL_MS = 5000;
const IDLE_INTERVAL_MS = 30000;

/** An instance can hold many unmerged PRs; the sidebar shows the recent ones. */
const MAX_ROWS = 8;

interface Row {
  key: string;
  issue: WorktreeIssue | null;
  kinds: WorktreeKind[];
  relPath: string;
  live: boolean;
  orphaned: boolean;
}

function toRow(wt: Worktree, seen: Set<number>): Row | null {
  const issue = wt.issue;
  if (issue) {
    if (seen.has(issue.number)) return null;
    seen.add(issue.number);
  }
  return {
    key: issue ? `issue:${issue.number}` : `wt:${wt.rel_path}`,
    issue,
    kinds: [wt.kind],
    relPath: wt.rel_path,
    live: wt.live,
    // A directory with no run behind it survived a hard kill: the stage that
    // created it never reached its deferred cleanup.
    orphaned: wt.live && issue?.run_status !== 'running',
  };
}

/**
 * One list, not two. An issue usually appears in both the live directories and
 * the unmerged-work query, so it gets a single row; the dot says whether an
 * agent is in it right now. Work in flight leads, because that is what changes
 * minute to minute — an issue being worked on has no branch or PR yet, which is
 * exactly the window a database-only list cannot show.
 */
function toRows(data: WorktreesResponse): Row[] {
  const seen = new Set<number>();
  const rows: Row[] = [];

  for (const wt of data.live) {
    const row = toRow(wt, seen);
    if (row) rows.push(row);
  }
  for (const wt of data.open) {
    const row = toRow(wt, seen);
    if (!row) continue;
    // An issue can hold several worktrees at once — a review checkout beside
    // the work checkout, say — so fold its kinds into the row already shown.
    const existing = rows.find((r) => r.issue && row.issue && r.issue.number === row.issue.number);
    if (existing) {
      if (!existing.kinds.includes(wt.kind)) existing.kinds.push(wt.kind);
      existing.live = existing.live || wt.live;
      continue;
    }
    rows.push(row);
  }
  return rows;
}

function WorktreeRow({ row, onNavigate }: { row: Row; onNavigate?: () => void }) {
  const issue = row.issue;
  const kinds = row.kinds.map((k) => KIND_META[k] ?? KIND_META.unknown);
  const stage = issue ? stageMeta(issue.stage) : null;
  const dotColor = row.orphaned ? '#f87171' : kinds[0]?.color ?? '#94a3b8';

  const body = (
    <>
      <div className="flex items-center gap-2">
        <PulseDot
          color={dotColor}
          className={clsx('!h-1.5 !w-1.5 shrink-0', !row.live && 'opacity-40')}
        />
        <span className="shrink-0 text-[11px] tabular-nums text-ink-faint">
          {issue ? `#${issue.number}` : '—'}
        </span>
        <span
          className={clsx(
            'min-w-0 flex-1 truncate text-[13px]',
            row.live ? 'text-ink' : 'text-ink-dim',
          )}
        >
          {issue?.title ?? row.relPath}
        </span>
        {issue?.pr_number ? (
          <span className="shrink-0 text-[10px] tabular-nums text-ink-faint">
            #{issue.pr_number}
          </span>
        ) : null}
      </div>
      <div className="mt-0.5 flex items-center gap-1.5 overflow-hidden whitespace-nowrap pl-3.5 text-[11px] text-ink-faint">
        {kinds.map((kind) => (
          <span key={kind.label} style={{ color: kind.color }}>
            {kind.label}
          </span>
        ))}
        {kinds.length > 1 && <span>·</span>}
        {stage && (
          <>
            {kinds.length === 1 && <span>·</span>}
            <span style={{ color: stage.color }}>{stage.label}</span>
          </>
        )}
        {row.orphaned && (
          <>
            <span>·</span>
            <span className="text-danger">orphaned</span>
          </>
        )}
      </div>
    </>
  );

  if (!issue) {
    return <div className="rounded-lg px-3 py-1.5">{body}</div>;
  }
  return (
    <Link
      href={`/issues/${issue.number}`}
      onClick={onNavigate}
      className="block rounded-lg px-3 py-1.5 transition hover:bg-panel-2"
    >
      {body}
    </Link>
  );
}

/**
 * The worktrees this instance is working in, and the issues that still have
 * unmerged work. Both come from the instance's own directory listing joined
 * with its database, so the browser never touches the filesystem.
 */
export function WorktreePanel({ onNavigate }: { onNavigate?: () => void }) {
  const { active } = useInstances();
  const [busy, setBusy] = useState(false);
  const { data, error, loading } = useInstanceData<WorktreesResponse>(
    (base) => api.worktrees(base),
    busy ? BUSY_INTERVAL_MS : IDLE_INTERVAL_MS,
  );

  // An agent working in a worktree is the one thing here that changes minute
  // to minute, so the poll rate follows it.
  const liveCount = data?.live.length ?? 0;
  useEffect(() => setBusy(liveCount > 0), [liveCount]);

  if (!active) return null;

  const rows = data ? toRows(data) : [];
  const shown = rows.slice(0, MAX_ROWS);
  const hidden = rows.length - shown.length;

  return (
    <section className="mt-5 border-t border-edge pt-4">
      <div className="mb-1.5 flex items-center gap-2 px-3">
        <GitBranch size={13} className="shrink-0 text-ink-faint" />
        <p className="flex-1 text-[10px] font-semibold uppercase tracking-[0.13em] text-ink-faint">
          Worktrees
        </p>
        {rows.length > 0 && (
          <span className="text-[10px] tabular-nums text-ink-faint">{rows.length}</span>
        )}
      </div>

      {error && !data && <p className="px-3 py-1 text-[11px] text-ink-faint">Unavailable.</p>}

      {!data && loading && (
        <div className="space-y-2 px-3 py-1.5">
          <div className="skeleton h-8 w-full" />
          <div className="skeleton h-8 w-4/5" />
        </div>
      )}

      {data && rows.length === 0 && (
        <p className="px-3 py-1 text-[11px] text-ink-faint">Nothing in flight, none open.</p>
      )}

      {shown.map((row) => (
        <WorktreeRow key={row.key} row={row} onNavigate={onNavigate} />
      ))}

      {hidden > 0 && <p className="px-3 pt-1.5 text-[11px] text-ink-faint">and {hidden} more</p>}
    </section>
  );
}
