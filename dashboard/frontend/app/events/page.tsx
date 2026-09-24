'use client';

import { useMemo, useState } from 'react';
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
import { api } from '@/lib/api';
import type { EventRecord } from '@/lib/types';
import { ActivityChart } from '@/components/activity-chart';
import { EventFeed } from '@/components/event-feed';
import { Badge, EmptyState, ErrorState, Panel, Skeleton, StatCard } from '@/components/ui';

const RANGES = [
  { value: 6, label: '6h' },
  { value: 24, label: '24h' },
  { value: 72, label: '3d' },
  { value: 168, label: '7d' },
] as const;

type EventFilter = 'all' | 'failures' | 'scheduler' | 'agents';

function eventMatchesFilter(event: EventRecord, filter: EventFilter): boolean {
  if (filter === 'all') return true;
  if (filter === 'failures') {
    return /fail|error|abort|max_iterations|unclassified/i.test(event.event_type);
  }
  if (filter === 'scheduler') return event.event_type.startsWith('tick.');
  return /spec\.|work\.|review\.|fix\.|approve|revise|retry/i.test(event.event_type);
}

export default function EventsPage() {
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<EventFilter>('all');
  const [hours, setHours] = useState(24);
  const { data, error, refresh, loading } = useInstanceData((base) => api.events(base, 500), 10000);
  const { data: activity } = useInstanceData(
    (base) => api.activity(base, hours),
    30000,
    true,
    [hours],
  );

  const events = data?.events ?? [];
  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return events.filter((event) => {
      if (!eventMatchesFilter(event, filter)) return false;
      if (!normalized) return true;
      return [event.event_type, event.stage, event.message, String(event.issue_number), String(event.pr_number)]
        .join(' ')
        .toLowerCase()
        .includes(normalized);
    });
  }, [events, filter, query]);

  const failureCount = events.filter((event) => eventMatchesFilter(event, 'failures')).length;
  const issueCount = new Set(events.map((event) => event.issue_number).filter(Boolean)).size;
  const busiest = activity?.buckets.reduce<{ time: string; count: number } | null>(
    (best, bucket) =>
      bucket.count > 0 && (!best || bucket.count > best.count) ? bucket : best,
    null,
  );

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
        <div className="flex items-center gap-1 rounded-xl border border-edge bg-panel p-1">
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

      <section className="grid grid-cols-2 gap-3 lg:grid-cols-4" aria-label="Event metrics">
        <StatCard
          label="Recorded"
          value={data?.count ?? '—'}
          sub="events in this snapshot"
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
        {activity ? <ActivityChart data={activity.buckets} /> : <Skeleton className="h-[230px] w-full" />}
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
          <Panel>
            <EmptyState
              icon={<TimerReset size={19} />}
              title={events.length ? 'No matching events' : 'No events recorded yet'}
              hint={
                events.length
                  ? 'Adjust the search or event filter.'
                  : 'Scheduler ticks and agent activity will appear here as Sloper runs.'
              }
            />
          </Panel>
        ) : (
          <Panel
            title="Event log"
            action={
              <span className="text-[10px] text-ink-faint">
                {filtered.length} of {events.length}
              </span>
            }
            bodyClassName="p-2"
          >
            <EventFeed events={filtered} />
          </Panel>
        )}
      </section>
    </div>
  );
}
