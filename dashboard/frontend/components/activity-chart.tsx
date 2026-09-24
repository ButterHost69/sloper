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

export function ActivityChart({ data }: { data: ActivityBucket[] }) {
  const total = data.reduce((sum, bucket) => sum + bucket.count, 0);
  const peak = data.reduce<ActivityBucket | null>(
    (best, bucket) => (!best || bucket.count > best.count ? bucket : best),
    null,
  );

  return (
    <div className="relative h-[230px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 8, left: -24, bottom: 0 }}>
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
            tickFormatter={(value) =>
              new Date(value).toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })
            }
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
          />
        </AreaChart>
      </ResponsiveContainer>
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
          Peak {peak.count} at {new Date(peak.time).toLocaleTimeString([], { hour: 'numeric' })}
        </div>
      )}
    </div>
  );
}
