'use client';

import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { ArrowRight } from 'lucide-react';
import { PIPELINE_ORDER, stageMeta } from '@/lib/stages';
import type { Summary } from '@/lib/types';

/**
 * Horizontal pipeline flow. Each node is a stage with a count.
 * The line between nodes is dimmed once the count reaches zero.
 */
export function PipelineFunnel({ summary }: { summary: Summary }) {
  const byStage = summary?.issues?.by_stage ?? {};
  const total = Math.max(1, summary?.issues?.total ?? 1);
  const router = useRouter();

  return (
    <div className="flex items-stretch gap-1.5 overflow-x-auto pb-1">
      {PIPELINE_ORDER.map((key, i) => {
        const meta = stageMeta(key);
        const count = byStage[key] ?? 0;
        const pct = Math.round((count / total) * 100);
        const isLast = i === PIPELINE_ORDER.length - 1;
        return (
          <div key={key} className="flex min-w-[92px] flex-1 items-center gap-1.5">
            <button
              onClick={() => router.push(`/issues?stage=${key}`)}
              className="group flex-1 rounded-lg border border-edge bg-panel-2 p-2.5 text-left transition hover:border-edge-2 hover:bg-white/[0.03]"
            >
              <div className="flex items-center justify-between">
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ background: meta.color, boxShadow: `0 0 6px ${meta.color}` }}
                />
                <span className="text-lg font-semibold tabular-nums text-ink">{count}</span>
              </div>
              <p className="mt-1.5 truncate text-[0.68rem] font-medium text-ink-dim">
                {meta.label}
              </p>
              <div className="mt-1.5 h-1 overflow-hidden rounded-full bg-edge">
                <div
                  className="h-full rounded-full transition-all duration-500"
                  style={{ width: `${pct}%`, background: meta.color, opacity: 0.75 }}
                />
              </div>
            </button>
            {!isLast && (
              <ArrowRight
                size={13}
                className={clsx('shrink-0', count === 0 ? 'text-edge-2' : 'text-ink-faint')}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}