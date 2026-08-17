'use client';

import { useMemo, useState } from 'react';
import { Search } from 'lucide-react';
import clsx from 'clsx';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { ErrorState, Panel } from '@/components/ui';
import { EventFeed } from '@/components/event-feed';

export default function EventsPage() {
  const [query, setQuery] = useState('');
  const [type, setType] = useState('all');
  const { data, error, refresh } = useInstanceData(
    (base) => api.events(base, 1000),
    10000,
  );

  const filtered = useMemo(() => {
    const events = data?.events ?? [];
    const q = query.trim().toLowerCase();
    return events.filter((e) => {
      if (type !== 'all' && e.event_type !== type) return false;
      if (!q) return true;
      return (
        e.event_type.toLowerCase().includes(q) ||
        e.message.toLowerCase().includes(q) ||
        String(e.issue_number).includes(q) ||
        String(e.pr_number).includes(q)
      );
    });
  }, [data, query, type]);

  const typeCounts = useMemo(() => {
    const c: Record<string, number> = {};
    for (const e of data?.events ?? []) c[e.event_type] = (c[e.event_type] ?? 0) + 1;
    return Object.entries(c).sort((a, b) => b[1] - a[1]);
  }, [data]);

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-ink">Events</h1>
          <p className="mt-1 text-xs text-ink-faint">
            {data?.count ?? 0} audit log entries · {filtered.length} shown
          </p>
        </div>
        <div className="relative w-full max-w-xs">
          <Search size={14} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-ink-faint" />
          <input
            className="input !pl-9"
            placeholder="Search events…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>
      </div>

      {error && !data ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : !data ? (
        <Panel className="p-0">
          <div className="space-y-3 p-4">
            {Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="skeleton h-10 w-full" />
            ))}
          </div>
        </Panel>
      ) : (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
          {/* Type filter sidebar */}
          <div className="lg:col-span-1">
            <Panel bodyClassName="p-2">
              <button
                onClick={() => setType('all')}
                className={clsx(
                  'flex w-full items-center justify-between rounded-md px-3 py-2 text-left text-sm transition',
                  type === 'all' ? 'bg-accent/10 text-accent' : 'hover:bg-white/[0.04] text-ink-dim',
                )}
              >
                <span>All types</span>
                <span className="font-mono text-xs text-ink-faint">{data?.events.length ?? 0}</span>
              </button>
              <div className="mt-1 max-h-[60vh] overflow-auto">
                {typeCounts.map(([t, n]) => (
                  <button
                    key={t}
                    onClick={() => setType(t)}
                    className={clsx(
                      'flex w-full items-center justify-between gap-2 rounded-md px-3 py-2 text-left text-sm transition',
                      type === t ? 'bg-accent/10 text-accent' : 'hover:bg-white/[0.04] text-ink-dim',
                    )}
                  >
                    <span className="truncate font-mono text-xs">{t}</span>
                    <span className="font-mono text-xs text-ink-faint">{n}</span>
                  </button>
                ))}
              </div>
            </Panel>
          </div>

          {/* Feed */}
          <div className="lg:col-span-3">
            <Panel bodyClassName="p-2">
              <EventFeed events={filtered} />
            </Panel>
          </div>
        </div>
      )}
    </div>
  );
}