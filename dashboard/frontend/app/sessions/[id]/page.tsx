'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import clsx from 'clsx';
import {
  ArrowLeft,
  ExternalLink,
  GitPullRequest,
  Info,
  Lock,
  Radio,
  RefreshCw,
} from 'lucide-react';
import { useSessionStream } from '@/hooks/use-session-stream';
import { SessionTranscript } from '@/components/session-transcript';
import { RunStageBadge, RunStatusBadge } from '@/components/badges';
import { Badge, EmptyState, ErrorState, Panel, Skeleton } from '@/components/ui';
import { formatBytes, formatTime, timeAgo } from '@/lib/format';
import { describeCurrent, formatCost, formatElapsed, formatTokens, currentStepMs } from '@/lib/sessions';

export default function SessionPage() {
  const params = useParams<{ id: string }>();
  const id = decodeURIComponent(String(params?.id ?? ''));
  const stream = useSessionStream(id);
  const { session, entries, live, modifiedAt, loading, gone, error, hasMoreBefore, loadingEarlier } = stream;

  // Re-render the "running for 2m 10s" line once a second.
  const [, setTick] = useState(0);
  useEffect(() => {
    if (!live) return;
    const timer = setInterval(() => setTick((value) => value + 1), 1000);
    return () => clearInterval(timer);
  }, [live]);

  if (loading) {
    return (
      <div className="space-y-5">
        <Skeleton className="h-10 w-72" />
        <Skeleton className="h-[520px] w-full" />
      </div>
    );
  }

  if (gone) {
    return (
      <div className="space-y-5">
        <BackLink />
        <Panel bodyClassName="p-0">
          <EmptyState
            title="This session is no longer on disk"
            hint="Sloper deletes pi session files when a plan is approved (the spec session), when a run is aborted, or when the pull request closes."
            action={
              <Link href="/sessions" className="btn mt-1">
                Back to sessions
              </Link>
            }
          />
        </Panel>
      </div>
    );
  }

  if (error && !session) {
    return (
      <div className="space-y-5">
        <BackLink />
        <Panel>
          <ErrorState error={error} onRetry={stream.reload} />
        </Panel>
      </div>
    );
  }

  const activity = session ? describeCurrent(session) : null;
  const elapsed = session && live ? currentStepMs(session) : null;

  return (
    <div className="space-y-4">
      <BackLink />

      <header className="flex flex-col gap-3 border-b border-edge/70 pb-4">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
          {session?.stage && <RunStageBadge stage={session.stage} />}
          {live ? (
            <Badge color="#34d399" dot>
              live
            </Badge>
          ) : (
            <Badge color="#94a3b8" dot>
              finished
            </Badge>
          )}
          <h1 className="min-w-0 truncate font-mono text-lg font-semibold tracking-tight text-ink">{id}</h1>
          {session && session.issue_number > 0 && (
            <Link href={`/issues/${session.issue_number}`} className="text-xs text-accent hover:underline">
              issue #{session.issue_number}
            </Link>
          )}
          {session?.pr && (
            <span className="chip" style={{ color: session.pr.state === 'merged' ? '#a78bfa' : '#38bdf8' }}>
              <GitPullRequest size={11} aria-hidden="true" />
              {session.pr.url ? (
                <a href={session.pr.url} target="_blank" rel="noreferrer" className="hover:underline">
                  PR #{session.pr.number} · {session.pr.state}
                </a>
              ) : (
                <>
                  PR #{session.pr.number} · {session.pr.state}
                </>
              )}
            </span>
          )}
          {session?.run && (
            <>
              <RunStatusBadge status={session.run.status} />
              <span className="text-[11px] tabular-nums text-ink-faint">
                run #{session.run.id} · {formatElapsed(session.run.duration_ms ?? 0)}
              </span>
            </>
          )}
          <button
            type="button"
            onClick={stream.reload}
            className="btn btn-ghost ml-auto !h-8 !min-h-8 !px-2.5"
            aria-label="Reload the transcript"
          >
            <RefreshCw size={13} />
          </button>
        </div>

        {session && (
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-[11px] text-ink-faint">
            {activity && (activity.tone === 'live' || activity.tone === 'error') && (
              <span
                className={clsx(
                  'inline-flex items-center gap-1.5 font-medium',
                  activity.tone === 'error' ? 'text-danger' : 'text-emerald',
                )}
              >
                {activity.tone === 'live' ? (
                  <Radio size={12} aria-hidden="true" />
                ) : (
                  <Info size={12} aria-hidden="true" />
                )}
                {activity.label}
                {activity.detail ? ` · ${activity.detail}` : ''}
                {elapsed !== null ? ` · ${formatElapsed(elapsed)}` : ''}
              </span>
            )}
            {session.summary.model && (
              <span>
                {session.summary.provider ? `${session.summary.provider}/` : ''}
                {session.summary.model}
              </span>
            )}
            <span>{session.summary.messages} messages</span>
            <span>
              {session.summary.tool_calls} tools
              {session.summary.tool_errors > 0 ? ` · ${session.summary.tool_errors} failed` : ''}
            </span>
            <span>
              {formatTokens(session.summary.total_tokens)} tokens · {formatCost(session.summary.cost_usd)}
            </span>
            <span>{formatBytes(session.size_bytes)}</span>
            <span>
              updated <time dateTime={modifiedAt}>{timeAgo(modifiedAt)}</time>
            </span>
            {session.summary.started_at && <span>started {formatTime(session.summary.started_at)}</span>}
          </div>
        )}

        {session?.run?.error_message && (
          <p className="rounded-lg border border-danger/30 bg-danger/[0.06] px-3 py-2 text-xs text-danger">
            {session.run.error_message}
          </p>
        )}
      </header>

      <Panel bodyClassName="p-0">
        <SessionTranscript
          entries={entries}
          live={live}
          hasMoreBefore={hasMoreBefore}
          loadingEarlier={loadingEarlier}
          onLoadEarlier={stream.loadEarlier}
          emptyHint={
            live
              ? 'pi has not written a message yet. The session file appears once the first model response completes.'
              : 'This session file has no messages.'
          }
        />
        <footer className="flex items-center gap-2 border-t border-edge px-4 py-2.5 text-[11px] text-ink-faint">
          <Lock size={11} aria-hidden="true" />
          Read-only view. Sloper never sends input to a worker from the console.
          <span className="ml-auto hidden tabular-nums sm:inline">
            {entries.length} entries · tailing {session?.file ?? id}
          </span>
        </footer>
      </Panel>
    </div>
  );
}

function BackLink() {
  return (
    <Link href="/sessions" className="inline-flex items-center gap-1.5 text-xs text-ink-dim transition hover:text-ink">
      <ArrowLeft size={13} aria-hidden="true" /> All sessions
      <ExternalLink size={11} className="text-ink-faint" aria-hidden="true" />
    </Link>
  );
}
