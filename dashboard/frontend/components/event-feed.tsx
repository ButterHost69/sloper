'use client';

import Link from 'next/link';
import {
  CheckCircle2,
  CircleDot,
  GitMerge,
  GitPullRequest,
  ListChecks,
  MessageSquare,
  RefreshCcw,
  Sparkles,
  TriangleAlert,
  Workflow,
  XCircle,
} from 'lucide-react';
import type { EventRecord } from '@/lib/types';
import { formatTime, timeAgo, truncate } from '@/lib/format';
import { EmptyState } from '@/components/ui';

const EVENT_STYLE: Record<
  string,
  { color: string; icon: typeof CircleDot }
> = {
  'tick.started': { color: '#94a3b8', icon: RefreshCcw },
  'tick.completed': { color: '#94a3b8', icon: CheckCircle2 },
  'tick.failed': { color: '#f87171', icon: XCircle },
  'spec.started': { color: '#38bdf8', icon: Sparkles },
  'spec.completed': { color: '#38bdf8', icon: CheckCircle2 },
  'spec.failed': { color: '#f87171', icon: XCircle },
  'spec.empty_output': { color: '#fbbf24', icon: TriangleAlert },
  'spec.replyComment': { color: '#38bdf8', icon: MessageSquare },
  'spec.unclassified_response': { color: '#fbbf24', icon: TriangleAlert },
  'approve.received': { color: '#a78bfa', icon: ListChecks },
  'revise.received': { color: '#fbbf24', icon: RefreshCcw },
  'abort.received': { color: '#f87171', icon: XCircle },
  'retry.received': { color: '#fbbf24', icon: RefreshCcw },
  'work.started': { color: '#a78bfa', icon: Workflow },
  'work.completed': { color: '#a78bfa', icon: CheckCircle2 },
  'work.failed': { color: '#f87171', icon: XCircle },
  'review.started': { color: '#e879f9', icon: GitPullRequest },
  'review.approved': { color: '#e879f9', icon: CheckCircle2 },
  'review.changes_requested': { color: '#fbbf24', icon: TriangleAlert },
  'review.failed': { color: '#f87171', icon: XCircle },
  'review.max_iterations': { color: '#fbbf24', icon: TriangleAlert },
  'fix.started': { color: '#fbbf24', icon: Workflow },
  'fix.completed': { color: '#fbbf24', icon: CheckCircle2 },
  'fix.failed': { color: '#f87171', icon: XCircle },
  'cleanup.sessions_deleted': { color: '#34d399', icon: GitMerge },
  'recovery.interrupted_run': { color: '#fbbf24', icon: RefreshCcw },
};

function styleFor(type: string) {
  return (
    EVENT_STYLE[type] ?? {
      color: '#71717a',
      icon: CircleDot,
    }
  );
}

export function EventRow({ ev }: { ev: EventRecord }) {
  const { color, icon: Icon } = styleFor(ev.event_type);
  return (
    <div className="group flex items-start gap-3 rounded-lg px-2 py-2 transition hover:bg-white/[0.03]">
      <span
        className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md border"
        style={{ color, borderColor: `${color}33`, background: `${color}12` }}
      >
        <Icon size={13} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate font-mono text-[0.7rem] font-medium" style={{ color }}>
            {ev.event_type}
          </span>
          {ev.stage && (
            <span className="chip !text-[0.6rem] uppercase text-ink-faint">{ev.stage}</span>
          )}
          {ev.issue_number > 0 && (
            <Link
              href={`/issues/${ev.issue_number}`}
              className="chip !text-[0.6rem] text-ink-dim transition hover:text-accent"
            >
              #{ev.issue_number}
            </Link>
          )}
          {ev.pr_number > 0 && (
            <span className="chip !text-[0.6rem] text-ink-dim">PR #{ev.pr_number}</span>
          )}
        </div>
        {ev.message && (
          <p className="mt-0.5 truncate text-xs text-ink-dim" title={ev.message}>
            {truncate(ev.message, 140)}
          </p>
        )}
      </div>
      <div className="shrink-0 text-right">
        <p className="text-[0.68rem] text-ink-faint" title={formatTime(ev.created_at)}>
          {timeAgo(ev.created_at)}
        </p>
      </div>
    </div>
  );
}

export function EventFeed({ events, compact }: { events: EventRecord[]; compact?: boolean }) {
  if (!events?.length) {
    return (
      <EmptyState title="No events yet" hint="Sloper activity will appear here once it starts ticking." />
    );
  }
  const list = compact ? events.slice(0, 12) : events;
  return (
    <div className="space-y-0.5">
      {list.map((ev) => (
        <EventRow key={ev.id} ev={ev} />
      ))}
    </div>
  );
}