'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import clsx from 'clsx';
import { Filter, Radio, Search } from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { EmptyState, ErrorState, Panel, StaleDataNotice } from '@/components/ui';
import { useInstances } from '@/components/instance-context';
import { PageHeader } from '@/components/page-header';
import { SessionsView } from '@/components/sessions-view';
import type { Session, SessionGroup } from '@/lib/types';

const FILTERS = ['all', 'live', 'finished'] as const;
type Filter = (typeof FILTERS)[number];

const FILTER_LABELS: Record<Filter, string> = {
  all: 'All',
  live: 'Live',
  finished: 'Finished',
};

function matchesFilter(session: Session, filter: Filter): boolean {
  if (filter === 'live') return session.live;
  if (filter === 'finished') return !session.live;
  return true;
}

function matchesQuery(session: Session, query: string): boolean {
  if (!query) return true;
  const needle = query.toLowerCase();
  return (
    String(session.issue_number).includes(needle) ||
    session.id.toLowerCase().includes(needle) ||
    (session.issue_title ?? '').toLowerCase().includes(needle) ||
    (session.summary.title ?? '').toLowerCase().includes(needle) ||
    (session.summary.model ?? '').toLowerCase().includes(needle) ||
    (session.stage ?? '').toLowerCase().includes(needle)
  );
}

export default function SessionsPage() {
  const [filter, setFilter] = useState<Filter>('all');
  const [groupBy, setGroupBy] = useState<'issue' | 'pr'>('issue');
  const [query, setQuery] = useState('');
  const { active } = useInstances();
  const { data, error, refresh } = useInstanceData((base) => api.sessions(base), 5000);

  const sessions = data?.sessions ?? [];

  const counts = useMemo(() => {
    const live = sessions.filter((session) => session.live).length;
    return { all: sessions.length, live, finished: sessions.length - live };
  }, [sessions]);

  const groups = useMemo<SessionGroup[]>(() => {
    const source = data?.groups ?? [];
    const trimmed = query.trim();
    return source
      .map((group) => ({
        ...group,
        sessions: group.sessions.filter(
          (session) => matchesFilter(session, filter) && matchesQuery(session, trimmed),
        ),
      }))
      .filter((group) => group.sessions.length > 0)
      .map((group) => ({
        ...group,
        live_count: group.sessions.filter((session) => session.live).length,
      }));
  }, [data, filter, query]);

  const shown = groups.reduce((total, group) => total + group.sessions.length, 0);

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Sessions"
        title="Agent sessions"
        subtitle={
          data
            ? `${shown} shown · ${counts.all} transcripts · ${counts.live} live right now`
            : 'Live transcripts of what each pi worker is doing'
        }
        actions={
          data ? (
            <div className="flex items-center gap-1 rounded-full border border-edge bg-panel-2 p-0.5">
              {(['issue', 'pr'] as const).map((mode) => (
                <button
                  key={mode}
                  type="button"
                  onClick={() => setGroupBy(mode)}
                  aria-pressed={groupBy === mode}
                  className={clsx(
                    'rounded-full px-3 py-1 text-[11px] font-medium transition',
                    groupBy === mode ? 'bg-accent/10 text-accent' : 'text-ink-dim hover:text-ink',
                  )}
                >
                  {mode === 'issue' ? 'By issue' : 'By pull request'}
                </button>
              ))}
            </div>
          ) : undefined
        }
      />

      <div className="space-y-3 rounded-xl border border-edge bg-panel p-3">
        <div className="flex flex-wrap items-center gap-2">
          <Filter size={13} className="mr-1 text-ink-faint" aria-hidden="true" />
          {FILTERS.map((item) => (
            <button
              key={item}
              type="button"
              onClick={() => setFilter(item)}
              aria-pressed={filter === item}
              className={clsx(
                'chip transition',
                filter === item
                  ? '!border-accent/50 !bg-accent/10 !text-accent'
                  : 'text-ink-dim hover:border-edge-2 hover:text-ink',
              )}
            >
              {item === 'live' && <Radio size={11} aria-hidden="true" />}
              {FILTER_LABELS[item]}
              <span className="text-ink-faint">{counts[item]}</span>
            </button>
          ))}
          <label className="ml-auto flex min-h-8 min-w-[190px] flex-1 items-center gap-2 rounded-full border border-edge bg-panel-2 px-3 sm:flex-none">
            <Search size={13} className="shrink-0 text-ink-faint" aria-hidden="true" />
            <span className="sr-only">Filter sessions</span>
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Issue, stage or model"
              className="w-full bg-transparent text-xs text-ink outline-none placeholder:text-ink-faint"
            />
          </label>
        </div>
      </div>

      {error && data && <StaleDataNotice error={error} onRetry={refresh} label="session" />}

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        active ? (
          <div className="space-y-4">
            <div className="skeleton h-28 w-full" />
            <div className="skeleton h-28 w-full" />
          </div>
        ) : (
          <Panel bodyClassName="p-0">
            <EmptyState
              title="Connect an instance to watch sessions"
              hint="Transcripts are read from the pi session files the worker writes inside the container."
              action={
                <Link href="/instances" className="btn btn-primary mt-1">
                  Configure an instance
                </Link>
              }
            />
          </Panel>
        )
      ) : groups.length === 0 ? (
        <Panel bodyClassName="p-0">
          <EmptyState
            title={counts.all ? 'No sessions match these filters' : 'No sessions on disk yet'}
            hint={
              counts.all
                ? 'Try another filter or clear the search.'
                : 'pi writes a session file once its first response completes, so a worker that just started will appear shortly.'
            }
            action={
              counts.all ? (
                <button
                  type="button"
                  className="btn mt-1"
                  onClick={() => {
                    setFilter('all');
                    setQuery('');
                  }}
                >
                  Clear filters
                </button>
              ) : (
                <button type="button" className="btn mt-1" onClick={() => void refresh()}>
                  Refresh instance
                </button>
              )
            }
          />
        </Panel>
      ) : (
        <SessionsView groups={groups} groupBy={groupBy} />
      )}

      {data && (
        <p className="px-1 text-[11px] leading-5 text-ink-faint">
          Read from <span className="font-mono">{data.dir}</span> — the same files pi writes while a stage runs.
          Sloper deletes them when a plan is approved, a run is aborted, or the pull request closes.
        </p>
      )}
    </div>
  );
}
