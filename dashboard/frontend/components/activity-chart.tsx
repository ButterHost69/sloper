'use client';

import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import type { ActivityBucket } from '@/lib/types';

function formatRangeLabel(hours: number): string {
  return hours < 48 ? `${hours} hours` : `${hours / 24} days`;
}

export function ActivityChart({ data, hours = 24 }: { data: ActivityBucket[]; hours?: number }) {
  const total = data.reduce((sum, bucket) => sum + bucket.count, 0);
  const peak = data.reduce<ActivityBucket | null>(
    (best, bucket) => (!best || bucket.count > best.count ? bucket : best),
    null,
  );
  const longRange = hours > 48;
  const chartLabel = total
    ? `Activity chart: ${total} events over ${formatRangeLabel(hours)}. Peak ${peak?.count ?? 0} events at ${peak ? new Date(peak.time).toISOString() : 'an unknown time'}.`
    : `Activity chart: no events over ${formatRangeLabel(hours)}.`;

  const tickFormatter = (value: string) => {
    const date = new Date(value);
    return date.toLocaleString([], {
      weekday: longRange ? 'short' : undefined,
      hour: 'numeric',
      minute: longRange ? undefined : '2-digit',
    });
  };

  return (
    <div>
      <div
        className={`relative w-full ${total === 0 ? 'h-[180px] sm:h-[200px]' : 'h-[230px]'}`}
        role="img"
        aria-label={chartLabel}
      >
        <div className="h-full w-full" aria-hidden="true">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart
              data={data}
              margin={{ top: 10, right: 8, left: -24, bottom: 0 }}
              accessibilityLayer={false}
            >
              <defs>
                <linearGradient id="activity-fill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor="#7aaaff" stopOpacity={0.34} />
                  <stop offset="100%" stopColor="#7aaaff" stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <CartesianGrid stroke="var(--color-edge)" strokeDasharray="3 5" vertical={false} />
              <XAxis
                dataKey="time"
                axisLine={false}
                tickLine={false}
                minTickGap={32}
                tick={{ fill: 'var(--color-ink-faint)', fontSize: 10 }}
                tickFormatter={tickFormatter}
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                allowDecimals={false}
                width={38}
                tick={{ fill: 'var(--color-ink-faint)', fontSize: 10 }}
              />
              <Tooltip
                cursor={{ stroke: 'var(--color-edge-2)', strokeWidth: 1 }}
                contentStyle={{
                  background: 'var(--color-panel-2)',
                  border: '0.5px solid var(--color-edge-2)',
                  borderRadius: 10,
                  color: 'var(--color-ink)',
                  fontSize: 12,
                  boxShadow: 'var(--app-shadow)',
                }}
                labelStyle={{ color: 'var(--color-ink-faint)', marginBottom: 4 }}
                labelFormatter={(value) => new Date(String(value)).toLocaleString()}
                formatter={(value) => [`${Number(value)} events`, 'Activity']}
              />
              <Area
                type="monotone"
                dataKey="count"
                stroke="#7aaaff"
                strokeWidth={2}
                fill="url(#activity-fill)"
                activeDot={{ r: 4, fill: '#7aaaff', stroke: 'var(--color-panel)', strokeWidth: 2 }}
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
        {total === 0 && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="rounded-xl border border-edge bg-panel/90 px-3 py-2 text-center shadow-lg">
              <p className="text-xs font-medium text-ink-dim">No events in this window</p>
              <p className="mt-1 text-[10px] text-ink-faint">The timeline will fill as the coordinator runs.</p>
            </div>
          </div>
        )}
        {total > 0 && peak && (
          <div className="pointer-events-none absolute right-3 top-2 rounded-lg border border-edge bg-panel/90 px-2.5 py-1.5 text-[10px] text-ink-faint shadow-sm">
            Peak {peak.count} at {new Date(peak.time).toLocaleString([], { weekday: longRange ? 'short' : undefined, hour: 'numeric' })}
          </div>
        )}
      </div>

      <details className="mt-2 rounded-lg border border-edge bg-panel-2/50 px-3 py-2">
        <summary className="cursor-pointer text-[11px] font-medium text-ink-dim">
          View hourly data table
        </summary>
        <div
          className="mt-3 max-h-56 overflow-auto rounded-lg border border-edge bg-panel"
          tabIndex={0}
          role="region"
          aria-label="Hourly event counts"
        >
          <table className="w-full text-left text-[11px]">
            <caption className="sr-only">Hourly event counts for the selected range</caption>
            <thead className="sticky top-0 bg-panel-2 text-ink-faint">
              <tr>
                <th className="px-3 py-2 font-medium">Hour</th>
                <th className="px-3 py-2 text-right font-medium">Events</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-edge text-ink-dim">
              {data.map((bucket) => (
                <tr key={bucket.time}>
                  <td className="px-3 py-2 font-mono">{new Date(bucket.time).toLocaleString()}</td>
                  <td className="px-3 py-2 text-right tabular-nums">{bucket.count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>
    </div>
  );
}
