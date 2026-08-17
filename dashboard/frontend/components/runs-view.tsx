'use client';

import { useState } from 'react';
import Link from 'next/link';
import clsx from 'clsx';
import { Braces, ChevronDown, Terminal } from 'lucide-react';
import type { RunRecord } from '@/lib/types';
import { durationMs, formatTime, truncate } from '@/lib/format';
import { RunStageBadge, RunStatusBadge } from '@/components/badges';
import { EmptyState } from '@/components/ui';

function LogBlock({ title, content }: { title: string; content: string }) {
  if (!content) return null;
  return (
    <div className="mt-3">
      <p className="mb-1 flex items-center gap-1.5 text-[0.65rem] font-semibold uppercase tracking-wider text-ink-faint">
        <Terminal size={12} /> {title}
      </p>
      <pre className="max-h-72 overflow-auto rounded-md border border-edge bg-[#0c0c0f] p-3 font-mono text-[0.72rem] leading-relaxed text-[#c9c9d4] whitespace-pre-wrap">
        {content}
      </pre>
    </div>
  );
}

function RunRow({ run, defaultOpen }: { run: RunRecord; defaultOpen: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  const failed = run.status === 'failed';
  return (
    <div
      className={clsx(
        'border-b border-edge/60 transition',
        failed && 'bg-danger/[0.03]',
        open && 'bg-white/[0.02]',
      )}
    >
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-3 px-3 py-2.5 text-left hover:bg-white/[0.02]"
      >
        <ChevronDown
          size={14}
          className={clsx('shrink-0 text-ink-faint transition-transform', open && 'rotate-180')}
        />
        <span className="w-12 shrink-0 font-mono text-xs text-ink-faint">#{run.id}</span>
        <span className="w-16 shrink-0">
          <RunStageBadge stage={run.stage} />
        </span>
        <span className="w-24 shrink-0">
          <RunStatusBadge status={run.status} />
        </span>
        <Link
          href={`/issues/${run.issue_number}`}
          onClick={(e) => e.stopPropagation()}
          className="shrink-0 font-mono text-xs text-accent hover:underline"
        >
          #{run.issue_number}
        </Link>
        <span className="min-w-0 flex-1 truncate text-xs text-ink-faint">
          {run.error_message || '—'}
        </span>
        <span className="shrink-0 font-mono text-xs tabular-nums text-ink-dim">
          {durationMs(run.started_at, run.ended_at)}
        </span>
        <span className="hidden w-36 shrink-0 text-right font-mono text-[0.68rem] text-ink-faint sm:block">
          {formatTime(run.started_at)}
        </span>
      </button>
      {open && (
        <div className="px-4 pb-4 pl-14">
          {run.error_message && (
            <div className="rounded-md border border-danger/30 bg-danger/10 px-3 py-2 text-xs text-danger">
              {run.error_message}
            </div>
          )}
          <LogBlock title="Agent output" content={run.agent_output} />
          <LogBlock title="Agent thinking" content={run.agent_thinking} />
          <LogBlock title="Shell log" content={run.shell_log} />
          {!run.agent_output && !run.agent_thinking && !run.shell_log && (
            <p className="flex items-center gap-1.5 text-xs text-ink-faint">
              <Braces size={13} /> No captured output for this run.
            </p>
          )}
        </div>
      )}
    </div>
  );
}

export function RunsView({ runs, defaultOpenFirst = false }: { runs: RunRecord[]; defaultOpenFirst?: boolean }) {
  if (!runs?.length) {
    return <EmptyState title="No runs yet" hint="Pipeline executions will show up here." />;
  }
  return (
    <div className="overflow-hidden rounded-lg border border-edge">
      <div className="flex items-center gap-3 border-b border-edge bg-panel-2 px-3 py-2 text-[0.65rem] font-semibold uppercase tracking-wider text-ink-faint">
        <span className="w-12">Run</span>
        <span className="w-16">Stage</span>
        <span className="w-24">Status</span>
        <span className="w-16">Issue</span>
        <span className="flex-1">Error / note</span>
        <span className="w-16 text-right">Duration</span>
        <span className="hidden w-36 text-right sm:block">Started</span>
      </div>
      {runs.map((r, i) => (
        <RunRow key={r.id} run={r} defaultOpen={defaultOpenFirst && i === 0} />
      ))}
    </div>
  );
}

export function truncateRun(run: RunRecord, n = 120) {
  return truncate(run.agent_output || run.error_message || '', n);
}