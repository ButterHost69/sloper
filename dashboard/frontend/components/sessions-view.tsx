'use client';

import Link from 'next/link';
import clsx from 'clsx';
import {
  ChevronRight,
  CircleAlert,
  Clock,
  GitPullRequest,
  Hash,
  MessageSquare,
  Radio,
  Wrench,
} from 'lucide-react';
import type { Session, SessionGroup } from '@/lib/types';
import { RunStageBadge, RunStatusBadge } from '@/components/badges';
import { Badge } from '@/components/ui';
import { formatBytes, timeAgo } from '@/lib/format';
import {
  describeCurrent,
  formatCost,
  formatElapsed,
  formatTokens,
  currentStepMs,
  groupSessionsByPR,
  type PRGroup,
} from '@/lib/sessions';

function StateChip({ state }: { state?: string }) {
  if (!state) return null;
  const color = state === 'open' ? '#34d399' : state === 'closed' ? '#94a3b8' : '#94a3b8';
  return <Badge color={color}>{state}</Badge>;
}

function PRChip({ session }: { session: Session }) {
  if (!session.pr) return null;
  const { pr } = session;
  const color = pr.state === 'merged' ? '#a78bfa' : pr.state === 'open' ? '#38bdf8' : '#94a3b8';
  const label = `PR #${pr.number} · ${pr.state}`;
  return pr.url ? (
    <a
      href={pr.url}
      target="_blank"
      rel="noreferrer"
      onClick={(event) => event.stopPropagation()}
      className="chip transition hover:border-edge-2"
      style={{ color }}
      aria-label={`${label} (opens on GitHub)`}
    >
      <GitPullRequest size={11} aria-hidden="true" /> {label}
    </a>
  ) : (
    <span className="chip" style={{ color }}>
      <GitPullRequest size={11} aria-hidden="true" /> {label}
    </span>
  );
}

function ActivityLine({ session }: { session: Session }) {
  const activity = describeCurrent(session);
  const elapsed = session.live ? currentStepMs(session) : null;
  const Icon = activity.tone === 'error'
    ? CircleAlert
    : activity.tone === 'live'
      ? Radio
      : Clock;

  return (
    <span className="flex min-w-0 items-center gap-1.5">
      <Icon
        size={13}
        className={clsx(
          'shrink-0',
          activity.tone === 'error'
            ? 'text-danger'
            : activity.tone === 'live'
              ? 'text-emerald'
              : 'text-ink-faint',
        )}
        aria-hidden="true"
      />
      <span
        className={clsx(
          'truncate text-xs font-medium',
          activity.tone === 'error' ? 'text-danger' : 'text-ink',
        )}
      >
        {activity.label}
      </span>
      {activity.detail && (
        <span className="hidden min-w-0 truncate text-[11px] text-ink-faint sm:inline">
          {activity.detail}
        </span>
      )}
      {elapsed !== null && (
        <span className="shrink-0 text-[11px] tabular-nums text-ink-faint">
          {formatElapsed(elapsed)}
        </span>
      )}
    </span>
  );
}

function SessionRow({ session }: { session: Session }) {
  return (
    <Link
      href={`/sessions/${encodeURIComponent(session.id)}`}
      className="group grid grid-cols-[auto_1fr_auto] items-center gap-x-3 gap-y-1 px-4 py-3 transition hover:bg-panel-2/50 sm:grid-cols-[auto_minmax(0,1fr)_auto_auto]"
      aria-label={`Open the ${session.stage || 'agent'} session for issue ${session.issue_number}`}
    >
      <span className="flex items-center gap-2">
        {session.live ? (
          <span className="live-dot h-2 w-2 rounded-full bg-emerald" aria-hidden="true" />
        ) : (
          <span className="h-2 w-2 rounded-full bg-edge-2/50" aria-hidden="true" />
        )}
        {session.stage ? (
          <RunStageBadge stage={session.stage} />
        ) : (
          <Badge color="#94a3b8">session</Badge>
        )}
        {session.fix_iteration ? (
          <span className="text-[10px] font-medium text-ink-faint">#{session.fix_iteration}</span>
        ) : null}
      </span>

      <span className="col-span-2 flex min-w-0 flex-col gap-0.5 sm:col-span-1">
        <ActivityLine session={session} />
        <span className="flex min-w-0 items-center gap-2 text-[11px] text-ink-faint">
          {session.run && <RunStatusBadge status={session.run.status} />}
          <span className="truncate">
            {session.summary.title || session.summary.last_text || session.file}
          </span>
        </span>
      </span>

      <span className="hidden items-center gap-3 text-[11px] tabular-nums text-ink-faint md:flex">
        <span className="inline-flex items-center gap-1" title="Messages">
          <MessageSquare size={11} aria-hidden="true" /> {session.summary.messages}
        </span>
        <span className="inline-flex items-center gap-1" title="Tool calls">
          <Wrench size={11} aria-hidden="true" /> {session.summary.tool_calls}
          {session.summary.tool_errors > 0 && (
            <span className="text-danger">/{session.summary.tool_errors}</span>
          )}
        </span>
        <span title="Tokens / cost">
          {formatTokens(session.summary.total_tokens)} · {formatCost(session.summary.cost_usd)}
        </span>
        <span className="hidden w-14 text-right lg:inline" title="Transcript size">
          {formatBytes(session.size_bytes)}
        </span>
      </span>

      <span className="flex items-center gap-2 text-[11px] text-ink-faint">
        <time dateTime={session.modified_at}>{timeAgo(session.modified_at)}</time>
        <ChevronRight
          size={14}
          className="text-ink-faint transition group-hover:translate-x-0.5 group-hover:text-ink"
          aria-hidden="true"
        />
      </span>
    </Link>
  );
}

function GroupCard({
  leading,
  title,
  subtitle,
  chips,
  liveCount,
  updatedAt,
  sessions,
}: {
  leading: React.ReactNode;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  chips?: React.ReactNode;
  liveCount: number;
  updatedAt?: string;
  sessions: Session[];
}) {
  return (
    <section className="panel overflow-hidden">
      <header className="flex flex-wrap items-center gap-x-3 gap-y-2 border-b border-edge px-4 py-3">
        <span className="flex items-center gap-2">{leading}</span>
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-center gap-2">
            <span className="truncate text-sm font-medium text-ink">{title}</span>
            {chips}
          </div>
          {subtitle && <div className="mt-0.5 flex items-center gap-2 text-[11px] text-ink-faint">{subtitle}</div>}
        </div>
        <div className="flex items-center gap-3 text-[11px] text-ink-faint">
          {liveCount > 0 && (
            <span className="inline-flex items-center gap-1.5 text-emerald">
              <span className="live-dot h-1.5 w-1.5 rounded-full bg-emerald" aria-hidden="true" />
              {liveCount} live
            </span>
          )}
          {updatedAt && <time dateTime={updatedAt}>{timeAgo(updatedAt)}</time>}
        </div>
      </header>
      <div className="divide-y divide-edge/60">
        {sessions.map((session) => (
          <SessionRow key={session.id} session={session} />
        ))}
      </div>
    </section>
  );
}

export function SessionsView({
  groups,
  groupBy,
}: {
  groups: SessionGroup[];
  groupBy: 'issue' | 'pr';
}) {
  if (groupBy === 'pr') {
    return (
      <div className="space-y-4">
        {groupSessionsByPR(groups).map((group: PRGroup) => (
          <GroupCard
            key={group.key}
            leading={
              group.pr ? (
                <GitPullRequest size={16} className="text-accent" aria-hidden="true" />
              ) : (
                <Hash size={16} className="text-ink-faint" aria-hidden="true" />
              )
            }
            title={
              group.pr ? (
                group.pr.url ? (
                  <a href={group.pr.url} target="_blank" rel="noreferrer" className="text-accent hover:underline">
                    PR #{group.pr.number}
                  </a>
                ) : (
                  <>PR #{group.pr.number}</>
                )
              ) : (
                group.title
              )
            }
            chips={
              <>
                {group.pr && <StateChip state={group.pr.state} />}
                {group.pr?.review_state && group.pr.review_state !== 'none' && (
                  <Badge color={group.pr.review_state === 'approved' ? '#34d399' : '#fbbf24'}>
                    {group.pr.review_state.replace(/_/g, ' ')}
                  </Badge>
                )}
              </>
            }
            subtitle={
              group.pr ? (
                <>
                  <span className="truncate">{group.title || 'No cached title'}</span>
                  {group.issueNumbers.map((number) => (
                    <Link key={number} href={`/issues/${number}`} className="text-accent hover:underline">
                      issue #{number}
                    </Link>
                  ))}
                </>
              ) : (
                <span>
                  {group.key === 'unscoped'
                    ? 'Sessions that do not belong to a sloper issue'
                    : 'Issues that have not produced a pull request yet'}
                </span>
              )
            }
            liveCount={group.live_count}
            updatedAt={group.updated_at}
            sessions={group.sessions}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {groups.map((group) => (
        <GroupCard
          key={group.issue_number}
          leading={
            group.unscoped ? (
              <Hash size={16} className="text-ink-faint" aria-hidden="true" />
            ) : (
              <Link
                href={`/issues/${group.issue_number}`}
                className="font-mono text-sm text-accent hover:underline"
              >
                #{group.issue_number}
              </Link>
            )
          }
          title={
            group.unscoped ? (
              group.title
            ) : (
              <Link href={`/issues/${group.issue_number}`} className="hover:underline">
                {group.title || `Issue #${group.issue_number}`}
              </Link>
            )
          }
          chips={
            <>
              {!group.unscoped && <StateChip state={group.state} />}
              {group.stage && <RunStageBadge stage={group.stage} />}
              {group.pr && (
                <span className="chip" style={{ color: group.pr.state === 'merged' ? '#a78bfa' : '#38bdf8' }}>
                  <GitPullRequest size={11} aria-hidden="true" /> #{group.pr.number} · {group.pr.state}
                </span>
              )}
            </>
          }
          subtitle={
            group.branch_name ? (
              <span className="truncate font-mono">{group.branch_name}</span>
            ) : undefined
          }
          liveCount={group.live_count}
          updatedAt={group.updated_at}
          sessions={group.sessions}
        />
      ))}
    </div>
  );
}
