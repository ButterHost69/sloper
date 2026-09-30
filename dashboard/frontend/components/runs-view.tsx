'use client';

import { Fragment, useState } from 'react';
import Link from 'next/link';
import clsx from 'clsx';
import { Braces, ChevronDown, Terminal } from 'lucide-react';
import type { RunRecord } from '@/lib/types';
import { durationMs, formatTime } from '@/lib/format';
import { RunStageBadge, RunStatusBadge } from '@/components/badges';
import { EmptyState } from '@/components/ui';

function LogBlock({ title, content }: { title: string; content: string }) {
  if (!content) return null;
  return (
    <div className="mt-3">
      <p className="mb-1 flex items-center gap-1.5 text-[0.65rem] font-semibold uppercase tracking-wider text-ink-faint">
        <Terminal size={12} /> {title}
      </p>
      <pre
        className="max-h-72 overflow-auto whitespace-pre-wrap rounded-lg border border-edge bg-base p-3 font-mono text-[0.72rem] leading-relaxed text-ink-dim"
        tabIndex={0}
        role="region"
        aria-label={`${title} log`}
      >
        {content}
      </pre>
    </div>
  );
}

function RunRow({ run, defaultOpen }: { run: RunRecord; defaultOpen: boolean }) {
  const [open, setOpen] = useState(defaultOpen);
  const failed = run.status === 'failed';
  const detailsId = `run-${run.id}-details`;

  return (
    <Fragment>
      <tr
        className={clsx(
          'border-b border-edge/60 transition',
          failed && 'bg-danger/[0.03]',
          open && 'bg-panel-2/35',
        )}
      >
        <td className="w-7 px-3 py-2.5 align-middle">
          <button
            type="button"
            onClick={() => setOpen((value) => !value)}
            className="flex h-7 w-7 items-center justify-center rounded-md text-ink-faint transition hover:bg-panel-2 hover:text-ink"
            aria-expanded={open}
            aria-controls={detailsId}
            aria-label={`${open ? 'Collapse' : 'Expand'} run ${run.id}`}
          >
            <ChevronDown
              size={14}
              className={clsx('shrink-0 transition-transform', open && 'rotate-180')}
              aria-hidden="true"
            />
          </button>
        </td>
        <td className="w-12 px-3 py-2.5 font-mono text-xs text-ink-faint">#{run.id}</td>
        <td className="w-16 px-3 py-2.5">
          <RunStageBadge stage={run.stage} />
        </td>
        <td className="w-24 px-3 py-2.5">
          <RunStatusBadge status={run.status} />
        </td>
        <td className="w-16 px-3 py-2.5">
          <Link
            href={`/issues/${run.issue_number}`}
            className="font-mono text-xs text-accent hover:underline"
            aria-label={`Open issue ${run.issue_number}`}
          >
            #{run.issue_number}
          </Link>
        </td>
        <td className="min-w-0 px-3 py-2.5 text-xs text-ink-faint">
          <div className="truncate">{run.error_message || '—'}</div>
        </td>
        <td className="w-16 px-3 py-2.5 text-right font-mono text-xs tabular-nums text-ink-dim">
          {durationMs(run.started_at, run.ended_at)}
        </td>
        <td className="hidden w-36 px-3 py-2.5 text-right font-mono text-[0.68rem] text-ink-faint sm:table-cell">
          {formatTime(run.started_at)}
        </td>
      </tr>
      {open && (
        <tr id={detailsId}>
          <td colSpan={8} className="px-4 pb-4 pl-14">
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
          </td>
        </tr>
      )}
    </Fragment>
  );
}

export function RunsView({ runs, defaultOpenFirst = false }: { runs: RunRecord[]; defaultOpenFirst?: boolean }) {
  if (!runs?.length) {
    return <EmptyState title="No runs yet" hint="Pipeline executions will show up here." />;
  }
  return (
    <div className="overflow-x-auto rounded-lg border border-edge">
      <table className="w-full min-w-[760px] border-collapse text-left" aria-label="Run history">
        <thead className="bg-panel-2 text-[0.65rem] font-semibold uppercase tracking-wider text-ink-faint">
          <tr>
            <th className="w-7 px-3 py-2"><span className="sr-only">Expand</span></th>
            <th className="w-12 px-3 py-2">Run</th>
            <th className="w-16 px-3 py-2">Stage</th>
            <th className="w-24 px-3 py-2">Status</th>
            <th className="w-16 px-3 py-2">Issue</th>
            <th className="px-3 py-2">Error / note</th>
            <th className="w-16 px-3 py-2 text-right">Duration</th>
            <th className="hidden w-36 px-3 py-2 text-right sm:table-cell">Started</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run, index) => (
            <RunRow key={run.id} run={run} defaultOpen={defaultOpenFirst && index === 0} />
          ))}
        </tbody>
      </table>
    </div>
  );
}
