'use client';

import clsx from 'clsx';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { CheckCircle2, TriangleAlert } from 'lucide-react';
import type { RunRecord, Summary } from '@/lib/types';
import { timeAgo } from '@/lib/format';
import { Skeleton } from '@/components/ui';

const FAILURE_STATUSES = new Set(['failed', 'interrupted']);

const STATUS_CLASSES: Record<string, string> = {
  failed: 'text-danger',
  interrupted: 'text-warn',
};

/**
 * Failures board — a layout distinct from the stat cards: a
 * divider-separated counter strip plus a compact feed of the most recent
 * failed/interrupted runs.
 */
export function FailuresPanel({
  summary,
  runs,
  loading,
}: {
  summary: Summary;
  runs: RunRecord[];
  loading?: boolean;
}) {
  const router = useRouter();
  const failedIssues = summary?.issues?.failed ?? 0;
  const failedRuns = summary?.runs?.failed ?? 0;
  const interrupted = summary?.runs?.interrupted ?? 0;
  const failures = (runs ?? []).filter((r) => FAILURE_STATUSES.has(r.status)).slice(0, 6);

  const counters = [
    { label: 'Failed issues', value: failedIssues, className: 'text-danger' },
    { label: 'Failed runs', value: failedRuns, className: 'text-danger' },
    { label: 'Interrupted', value: interrupted, className: 'text-warn' },
  ];

  return (
    <section className="overflow-hidden rounded-xl border border-danger/20 bg-danger/[0.03]">
      <header className="flex items-center justify-between border-b border-danger/15 px-4 py-3">
        <h3 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-ink-dim">
          <TriangleAlert size={13} className="text-danger" /> Failures
        </h3>
        <Link href="/runs?status=attention" className="hit-target text-xs text-accent hover:underline">
          View all →
        </Link>
      </header>

      {loading ? (
        <div className="space-y-3 p-4">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-8 w-full" />
        </div>
      ) : (
        <div>
          {/* Counter strip — deliberately not StatCard style */}
          <div className="grid grid-cols-3 divide-x divide-danger/10 border-b border-danger/10">
            {counters.map((c) => (
              <div key={c.label} className="px-4 py-3">
                <p className={clsx('text-2xl font-semibold tabular-nums', c.className)}>
                  {c.value}
                </p>
                <p className="mt-0.5 text-[0.65rem] font-medium uppercase tracking-wider text-ink-faint">
                  {c.label}
                </p>
              </div>
            ))}
          </div>

          {/* Latest failures feed */}
          {failures.length === 0 ? (
            <div className="flex items-center gap-2 px-4 py-5 text-xs text-emerald">
              <CheckCircle2 size={14} /> No recent failures — all clear.
            </div>
          ) : (
            <ul className="divide-y divide-edge/60">
              {failures.map((r) => {
                const statusClass = STATUS_CLASSES[r.status] ?? 'text-ink-dim';
                return (
                  <li key={r.id}>
                    <button
                      onClick={() => router.push(`/issues/${r.issue_number}`)}
                      className="flex w-full items-center gap-3 px-4 py-2 text-left transition hover:bg-panel-2/60"
                    >
                      <span className="w-12 shrink-0 font-mono text-xs font-semibold text-accent">
                        #{r.issue_number}
                      </span>
                      <span className="w-16 shrink-0 text-[0.65rem] font-medium uppercase tracking-wider text-ink-faint">
                        stage: {r.stage}
                      </span>
                      <span
                        className={clsx('min-w-0 flex-1 truncate text-xs', statusClass)}
                      >
                        {r.error_message || `run #${r.id}`}
                      </span>
                      <span className="shrink-0 text-[0.65rem] tabular-nums text-ink-faint">
                        {timeAgo(r.ended_at || r.started_at)}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      )}
    </section>
  );
}