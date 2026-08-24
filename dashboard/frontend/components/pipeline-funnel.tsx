'use client';

import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { ArrowRight } from 'lucide-react';
import { PIPELINE_ORDER, stageMeta } from '@/lib/stages';
import type { Summary } from '@/lib/types';

/**
 * Horizontal pipeline flow. Each node is a stage with a count.
 * spec-ongoing + spec-done and approved + work-done each share a single
 * card with a split counter.
 */

interface MergedStage {
  cardKey: string;
  mergedKey: string;
  label: string;
  /** Pinned ongoing count when the backend doesn't expose the stage yet. */
  fixedOngoing?: number;
}

const MERGED_STAGES: MergedStage[] = [
  { cardKey: 'spec-ongoing', mergedKey: 'spec-done', label: 'Spec' },
  { cardKey: 'approved', mergedKey: 'work-done', label: 'Working' },
  // Reviewing stage to be added in Sloper soon — until then the ongoing
  // counter is pinned at 0 and only review-done is shown.
  { cardKey: 'review-done', mergedKey: '', label: 'Review', fixedOngoing: 0 },
];

export function PipelineFunnel({ summary }: { summary: Summary }) {
  const byStage = summary?.issues?.by_stage ?? {};
  const failedCount = byStage['failed'] ?? 0;
  const router = useRouter();

  return (
    <div className="flex items-stretch gap-1.5 overflow-x-auto pb-1">
      {PIPELINE_ORDER.map((key, i) => {
        const merged = MERGED_STAGES.find(
          (m) => m.cardKey === key || (m.mergedKey !== '' && m.mergedKey === key),
        );
        if (merged && merged.mergedKey !== '' && key === merged.mergedKey) return null; // merged into the cardKey card
        const isCard = merged?.cardKey === key;
        const meta = isCard ? { ...stageMeta(key), label: merged!.label } : stageMeta(key);
        const ongoing = isCard
          ? (merged!.fixedOngoing ?? (byStage[merged!.cardKey] ?? 0))
          : 0;
        const done = isCard
          ? merged!.mergedKey !== ''
            ? (byStage[merged!.mergedKey] ?? 0)
            : (byStage[merged!.cardKey] ?? 0)
          : 0;
        const count = isCard ? ongoing + done : (byStage[key] ?? 0);
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
              {isCard ? (
                <p className="mt-0.5 truncate text-[0.62rem] tabular-nums text-ink-faint">
                  {ongoing} ongoing · {done} done
                </p>
              ) : (
                <p className="mt-0.5 text-[0.62rem]">&nbsp;</p>
              )}
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
      {failedCount > 0 && (
        <button
          onClick={() => router.push('/issues?stage=failed')}
          className="ml-2 shrink-0 self-stretch rounded-lg border border-danger/40 bg-danger/10 p-2.5 text-left transition hover:border-danger/60 hover:bg-danger/15"
          title="Issues stuck in a failed state"
        >
          <div className="flex items-center justify-between gap-3">
            <span
              className="h-2 w-2 rounded-full"
              style={{ background: '#f87171', boxShadow: '0 0 6px #f87171' }}
            />
            <span className="text-lg font-semibold tabular-nums text-danger">{failedCount}</span>
          </div>
          <p className="mt-1.5 text-[0.68rem] font-medium text-danger">Failed</p>
        </button>
      )}
    </div>
  );
}