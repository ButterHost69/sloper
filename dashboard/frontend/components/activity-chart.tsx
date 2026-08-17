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

export function ActivityChart({
  buckets,
  height = 220,
}: {
  buckets: ActivityBucket[];
  height?: number;
}) {
  const data = (buckets ?? []).map((b) => ({
    ...b,
    label: new Date(b.time).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    }),
  }));

  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: -18 }}>
        <defs>
          <linearGradient id="activityFill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#34d399" stopOpacity={0.35} />
            <stop offset="100%" stopColor="#34d399" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid stroke="#232329" strokeDasharray="3 3" vertical={false} />
        <XAxis
          dataKey="label"
          tick={{ fill: '#71717a', fontSize: 11 }}
          axisLine={{ stroke: '#232329' }}
          tickLine={false}
          interval="preserveStartEnd"
          minTickGap={40}
        />
        <YAxis
          allowDecimals={false}
          tick={{ fill: '#71717a', fontSize: 11 }}
          axisLine={false}
          tickLine={false}
        />
        <Tooltip
          contentStyle={{
            background: '#17171c',
            border: '1px solid #2e2e36',
            borderRadius: 8,
            fontSize: 12,
            color: '#e7e7ea',
          }}
          labelStyle={{ color: '#a1a1aa' }}
          cursor={{ stroke: '#2e2e36' }}
          formatter={(value) => [`${value} events`, 'Activity']}
        />
        <Area
          type="monotone"
          dataKey="count"
          stroke="#34d399"
          strokeWidth={2}
          fill="url(#activityFill)"
          dot={false}
          activeDot={{ r: 4, fill: '#34d399', stroke: '#09090b', strokeWidth: 2 }}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}