'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import {
  Activity,
  AlertCircle,
  CalendarClock,
  Filter,
  Hash,
  Search,
  TimerReset,
} from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { useInstances } from '@/components/instance-context';
import { api } from '@/lib/api';
import type { EventRecord } from '@/lib/types';
import { ActivityChart } from '@/components/activity-chart';
import { EventFeed } from '@/components/event-feed';
import { Badge, EmptyState, ErrorState, Panel, Skeleton, StaleDataNotice, StatCard } from '@/components/ui';

const RANGES = [
  { value: 6, label: '6h' },
  { value: 24, label: '24h' },
  { value: 72, label: '3d' },
  { value: 168, label: '7d' },
] as const;

type EventFilter = 'all' | 'failures' | 'scheduler' | 'agents';
const PAGE_SIZE = 500;

function eventMatchesFilter(event: EventRecord, filter: EventFilter): boolean {
  if (filter === 'all') return true;
  if (filter === 'failures') {
    return /fail|error|abort|max_iterations|unclassified|invalid|interrupt|no_changes/i.test(
      event.event_type,
    );
  }
  if (filter === 'scheduler') return event.event_type.startsWith('tick.');
  return /spec\.|work\.|review\.|fix\.|merge\.|approve|revise|retry/i.test(event.event_type);
}

function eventInRange(event: EventRecord, hours: number): boolean {
  const timestamp = Date.parse(event.created_at);
  if (Number.isNaN(timestamp)) return true;
  return timestamp >= Date.now() - hours * 60 * 60 * 1000;
}

function NoInstanceState() {
  return (
    <Panel bodyClassName="p-0">
      <EmptyState
        icon={<Activity size={19} />}
        title="Connect an instance to inspect events"
        hint="The event stream is read-only and becomes available when a Sloper API is configured."
        action={
          <Link href="/instances" className="btn btn-primary mt-1">
            Configure an instance
          </Link>
        }
      />
    </Panel>
  );
}

export default function EventsPage() {
  const { active } = useInstances();
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<EventFilter>('all');
  const [hours, setHours] = useState(24);
  const [olderEvents, setOlderEvents] = useState<EventRecord[]>([]);
  const [loadingMore, setLoadingMore] = useState(false);
  const [moreError, setMoreError] = useState<Error | null>(null);
  const [exhausted, setExhausted] = useState(false);
  const activeUrlRef = useRef(active?.url);
  const {
    data,
    error,
    refresh,
    loading,
  } = useInstanceData((base) => api.events(base, PAGE_SIZE), 10000);
  const {
    data: activity,
    error: activityError,
    refresh: refreshActivity,
    loading: activityLoading,
    refreshing: activityRefreshing,
  } = useInstanceData((base) => api.activity(base, hours), 30000, true, [hours]);

  useEffect(() => {
    activeUrlRef.current = active?.url;
    setOlderEvents([]);
    setMoreError(null);
    setExhausted(false);
    setLoadingMore(false);
  }, [active?.url]);

  const hasMore = !exhausted && (data?.events.length ?? 0) === PAGE_SIZE;

  const events = useMemo(() => {
    const byId = new Map<number, EventRecord>();
    for (const event of [...(data?.events ?? []), ...olderEvents]) byId.set(event.id, event);
    return [...byId.values()];
  }, [data, olderEvents]);

  const rangedEvents = useMemo(
    () => events.filter((event) => eventInRange(event, hours)),
    [events, hours],
  );

  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return rangedEvents.filter((event) => {
      if (!eventMatchesFilter(event, filter)) return false;
      if (!normalized) return true;
      return [event.event_type, event.stage, event.message, String(event.issue_number), String(event.pr_number)]
        .join(' ')
        .toLowerCase()
        .includes(normalized);
    });
  }, [filter, query, rangedEvents]);

  const failureCount = rangedEvents.filter((event) => eventMatchesFilter(event, 'failures')).length;
  const issueCount = new Set(rangedEvents.map((event) => event.issue_number).filter(Boolean)).size;
  const busiest = activity?.buckets.reduce<{ time: string; count: number } | null>(
    (best, bucket) =>
      bucket.count > 0 && (!best || bucket.count > best.count) ? bucket : best,
    null,
  );

  const loadMore = async () => {
    if (!active || loadingMore || !hasMore) return;
    const requestUrl = active.url;
    setLoadingMore(true);
    setMoreError(null);
    try {
      const cursor = events[events.length - 1]?.id;
      if (!cursor) return;
      const next = await api.events(requestUrl, PAGE_SIZE, 0, cursor);
      if (activeUrlRef.current !== requestUrl) return;
      setOlderEvents((current) => [...current, ...next.events]);
      setExhausted(next.events.length < PAGE_SIZE);
    } catch (loadError) {
      if (activeUrlRef.current === requestUrl) {
        setMoreError(loadError instanceof Error ? loadError : new Error('Could not load older events'));
      }
    } finally {
      if (activeUrlRef.current === requestUrl) setLoadingMore(false);
    }
  };

  const loadMoreControl =
    hasMore || loadingMore || moreError ? (
      <div className="flex flex-col items-center gap-2">
        {moreError && <p className="text-xs text-danger">{moreError.message}</p>}
        <button
          type="button"
          className="btn"
          onClick={() => void loadMore()}
          disabled={loadingMore || !hasMore}
        >
          {loadingMore ? 'Loading older events…' : 'Load older events'}
        </button>
      </div>
    ) : null;

  return (
    <div className="space-y-6">
      <section className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.14em] text-accent">
            <Activity size={12} /> Audit stream
          </div>
          <h1 className="mt-2 text-[28px] font-semibold tracking-[-0.04em] text-ink sm:text-[32px]">
            Events
          </h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-ink-dim">
            A read-only timeline of scheduler ticks, agent stages, review decisions, and recovery events.
          </p>
        </div>
        <div className="flex items-center gap-1 rounded-xl border border-edge bg-panel p-1" aria-label="Activity range">
          {RANGES.map((range) => (
            <button
              key={range.value}
              type="button"
              onClick={() => setHours(range.value)}
              className={`rounded-lg px-2.5 py-1.5 text-[10px] font-semibold transition ${
                hours === range.value
                  ? 'bg-accent/12 text-accent'
                  : 'text-ink-faint hover:bg-panel-2 hover:text-ink'
              }`}
              aria-pressed={hours === range.value}
            >
              {range.label}
            </button>
          ))}
        </div>
      </section>

      {!active ? (
        <NoInstanceState />
      ) : (
        <>
          <section className="grid grid-cols-2 gap-3 lg:grid-cols-4" aria-label="Event metrics">
            <StatCard
              label="Recorded"
              value={loading ? '—' : rangedEvents.length}
              sub={`events in the last ${hours < 48 ? `${hours} hours` : `${hours / 24} days`}`}
              icon={<Activity size={15} />}
              loading={loading}
            />
            <StatCard
              label="Failures"
              value={failureCount}
              sub="errors and warnings"
              icon={<AlertCircle size={15} />}
              accent="#f87171"
            />
            <StatCard
              label="Issues touched"
              value={issueCount}
              sub="unique issue numbers"
              icon={<Hash size={15} />}
              accent="#a78bfa"
            />
            <StatCard
              label="Busiest hour"
              value={busiest?.count ?? '—'}
              sub={busiest ? new Date(busiest.time).toLocaleString() : 'no activity yet'}
              icon={<CalendarClock size={15} />}
              accent="#4ed17e"
            />
          </section>

          <Panel
            title="Activity"
            action={
              <Badge color="#7aaaff" dot>
                Last {hours < 48 ? `${hours} hours` : `${hours / 24} days`}
              </Badge>
            }
          >
            {activityError ? (
              <ErrorState error={activityError} onRetry={refreshActivity} compact />
            ) : activityLoading || activityRefreshing ? (
              <Skeleton className="h-[230px] w-full" />
            ) : activity ? (
              <ActivityChart data={activity.buckets} hours={hours} />
            ) : (
              <Skeleton className="h-[230px] w-full" />
            )}
          </Panel>

          <section className="space-y-3">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div className="relative w-full lg:max-w-sm">
                <Search
                  size={14}
                  className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint"
                />
                <input
                  className="input !pl-9"
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Search event, issue, message…"
                  aria-label="Search events"
                />
              </div>
              <div className="flex flex-wrap items-center gap-1.5">
                <Filter size={13} className="mr-1 text-ink-faint" />
                {(
                  [
                    ['all', 'All'],
                    ['failures', 'Failures'],
                    ['scheduler', 'Scheduler'],
                    ['agents', 'Agent stages'],
                  ] as const
                ).map(([value, label]) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setFilter(value)}
                    className={`chip transition ${
                      filter === value
                        ? '!border-accent/45 !bg-accent/10 !text-accent'
                        : 'text-ink-dim hover:border-edge-2 hover:text-ink'
                    }`}
                    aria-pressed={filter === value}
                  >
                    {label}
                  </button>
                ))}
              </div>
            </div>

            {error && data && <StaleDataNotice error={error} onRetry={refresh} label="event" />}

            {error && !data ? (
              <Panel>
                <ErrorState error={error} onRetry={refresh} />
              </Panel>
            ) : !data ? (
              <Panel>
                <div className="space-y-2">
                  {Array.from({ length: 7 }).map((_, index) => (
                    <Skeleton key={index} className="h-12 w-full" />
                  ))}
                </div>
              </Panel>
            ) : filtered.length === 0 ? (
              <>
                <Panel bodyClassName="p-0">
                  <EmptyState
                    icon={<TimerReset size={19} />}
                    title={rangedEvents.length ? 'No matching events' : 'No events in this range'}
                    hint={
                      rangedEvents.length
                        ? 'Adjust the search or event filter.'
                        : 'Try a wider range or wait for the coordinator to record activity.'
                    }
                    action={
                      rangedEvents.length || hours >= 168 ? (
                        <button
                          type="button"
                          className="btn mt-1"
                          onClick={() => {
                            setQuery('');
                            setFilter('all');
                          }}
                        >
                          Clear filters
                        </button>
                      ) : (
                        <button type="button" className="btn mt-1" onClick={() => setHours(168)}>
                          View 7 days
                        </button>
                      )
                    }
                  />
                </Panel>
                {loadMoreControl}
              </>
            ) : (
              <>
                <Panel
                  title="Event log"
                  action={
                    <span className="text-[10px] text-ink-faint">
                      {filtered.length} of {rangedEvents.length} in range
                    </span>
                  }
                  bodyClassName="p-2"
                >
                  <EventFeed events={filtered} />
                </Panel>
                {loadMoreControl}
              </>
            )}
          </section>
        </>
      )}
    </div>
  );
}
