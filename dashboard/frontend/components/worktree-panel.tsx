'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { GitBranch } from 'lucide-react';
import clsx from 'clsx';
import { useInstances } from '@/components/instance-context';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { stageMeta } from '@/lib/stages';
import { PulseDot, Skeleton } from '@/components/ui';
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

interface Row {
  key: string;
  issue: WorktreeIssue | null;
  kinds: WorktreeKind[];
  relPath: string;
  live: boolean;
  orphaned: boolean;
}

/**
 * One issue can hold several worktrees at once — a review checkout alongside
 * the work checkout, say — so live entries collapse onto a single row per
 * issue. Directories sloper did not name keep a row of their own.
 */
function toRows(entries: Worktree[]): Row[] {
  const rows: Row[] = [];
  const byIssue = new Map<number, Row>();

  for (const wt of entries) {
    const issue = wt.issue;
    // A directory with no run behind it survived a hard kill: the stage that
    // created it never reached its deferred cleanup.
    const orphaned = wt.live && issue?.run_status !== 'running';

    if (!issue) {
      rows.push({
        key: `wt:${wt.rel_path}`,
        issue: null,
        kinds: [wt.kind],
        relPath: wt.rel_path,
        live: wt.live,
        orphaned,
      });
      continue;
    }

    const existing = byIssue.get(issue.number);
    if (existing) {
      if (!existing.kinds.includes(wt.kind)) existing.kinds.push(wt.kind);
      existing.orphaned = existing.orphaned && orphaned;
      continue;
    }
    const row: Row = {
      key: `issue:${issue.number}`,
      issue,
      kinds: [wt.kind],
      relPath: wt.rel_path,
      live: wt.live,
      orphaned,
    };
    byIssue.set(issue.number, row);
    rows.push(row);
  }
  return rows;
}

function GroupLabel({ children, count }: { children: React.ReactNode; count: number }) {
  return (
    <div className="flex items-center justify-between px-2.5 pb-1 pt-3">
      <span className="text-[0.6rem] font-semibold uppercase tracking-widest text-ink-faint">
        {children}
      </span>
      {count > 0 && <span className="text-[0.6rem] tabular-nums text-ink-faint">{count}</span>}
    </div>
  );
}

function WorktreeRow({ row, showLive }: { row: Row; showLive: boolean }) {
  const issue = row.issue;
  const stage = issue ? stageMeta(issue.stage) : null;
  const kinds = row.kinds.map((k) => KIND_META[k] ?? KIND_META.unknown);
  const dotColor = row.orphaned ? '#f87171' : kinds[0]?.color ?? '#94a3b8';

  const body = (
    <>
      <div className="flex items-center gap-2">
        {showLive && row.live && (
          <PulseDot color={dotColor} className="!h-1.5 !w-1.5 shrink-0" />
        )}
        <span className="shrink-0 text-[0.7rem] tabular-nums text-ink-faint">
          {issue ? `#${issue.number}` : '—'}
        </span>
        <span
          className={clsx(
            'min-w-0 flex-1 truncate text-xs',
            row.orphaned ? 'text-ink-dim' : 'text-ink',
          )}
        >
          {issue?.title ?? row.relPath}
        </span>
        {issue?.pr_number ? (
          <span className="shrink-0 text-[0.6rem] tabular-nums text-ink-faint">
            #{issue.pr_number}
          </span>
        ) : null}
      </div>
      <div className="mt-1 flex items-center gap-1.5 overflow-hidden whitespace-nowrap pl-3.5 text-[0.65rem] text-ink-faint">
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
    return <div className="rounded-lg px-2.5 py-2">{body}</div>;
  }
  return (
    <Link
      href={`/issues/${issue.number}`}
      className="block rounded-lg px-2.5 py-2 transition hover:bg-white/[0.04]"
    >
      {body}
    </Link>
  );
}

/**
 * The worktrees this instance is working in, plus the issues that still have
 * unmerged work. Both come from the instance's own directory listing joined
 * with its database, so the browser never touches the filesystem.
 */
export function WorktreePanel() {
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

  const liveRows = data ? toRows(data.live) : [];
  const openRows = data ? toRows(data.open) : [];
  const total = liveRows.length + openRows.length;

  return (
    <section className="border-t border-edge pt-1">
      <div className="flex items-center gap-2 px-3 pb-1 pt-4">
        <GitBranch size={13} className="text-ink-faint" />
        <span className="flex-1 text-[0.65rem] font-semibold uppercase tracking-widest text-ink-faint">
          Worktrees
        </span>
        {total > 0 && (
          <span className="rounded bg-white/[0.06] px-1.5 text-[0.65rem] tabular-nums text-ink-dim">
            {total}
          </span>
        )}
      </div>

      {error && !data && (
        <p className="px-3 py-1 text-[0.65rem] text-ink-faint">Worktrees unavailable.</p>
      )}

      {!data && loading && (
        <div className="space-y-2 px-3 py-2">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-4/5" />
        </div>
      )}

      {data && total === 0 && (
        <p className="px-3 py-1 text-[0.65rem] text-ink-faint">
          Nothing in flight, no open worktrees.
        </p>
      )}

      {liveRows.length > 0 && (
        <>
          <GroupLabel count={liveRows.length}>Working now</GroupLabel>
          {liveRows.map((row) => (
            <WorktreeRow key={row.key} row={row} showLive />
          ))}
        </>
      )}

      {openRows.length > 0 && (
        <>
          <GroupLabel count={openRows.length}>Open work</GroupLabel>
          {openRows.map((row) => (
            <WorktreeRow key={row.key} row={row} showLive />
          ))}
        </>
      )}
    </section>
  );
}
