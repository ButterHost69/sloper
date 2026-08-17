'use client';

import clsx from 'clsx';
import { AlertOctagon, Check, Loader2 } from 'lucide-react';
import { PIPELINE_ORDER, stageMeta } from '@/lib/stages';

export type DagNodeState = 'done' | 'current' | 'next' | 'failed' | 'skipped';

function computeStates(stage: string): Record<string, DagNodeState> {
  const states: Record<string, DagNodeState> = {};
  const idx = PIPELINE_ORDER.indexOf(stage as (typeof PIPELINE_ORDER)[number]);

  if (stage === 'failed') {
    // Mark the first non-empty ancestor as the failure point if possible;
    // otherwise flag the whole path red.
    PIPELINE_ORDER.forEach((key) => {
      states[key] = key === 'new' ? 'failed' : 'skipped';
    });
    states.failed = 'failed';
    return states;
  }

  PIPELINE_ORDER.forEach((key, i) => {
    if (idx === -1) {
      states[key] = 'next';
    } else if (i < idx) states[key] = 'done';
    else if (i === idx) states[key] = 'current';
    else states[key] = 'next';
  });

  // review-done acts as the terminal "waiting to merge" state.
  if (stage === 'review-done') {
    states['merged'] = 'next';
  }
  return states;
}

export function PipelineDAG({ stage }: { stage: string }) {
  const states = computeStates(stage);

  return (
    <div className="flex items-stretch gap-0 overflow-x-auto pb-1">
      {PIPELINE_ORDER.map((key, i) => {
        const meta = stageMeta(key);
        const state = states[key] ?? 'next';
        const isLast = i === PIPELINE_ORDER.length - 1;
        return (
          <div key={key} className="flex min-w-[86px] flex-1 items-center">
            <div className="flex-1">
              <div className="relative flex flex-col items-center gap-1.5">
                {/* node */}
                <div
                  className={clsx(
                    'relative flex h-9 w-9 items-center justify-center rounded-full border transition-all',
                    state === 'done' && 'border-transparent',
                    state === 'current' && 'border-transparent shadow-[0_0_16px]',
                    state === 'next' && 'border-edge-2',
                    state === 'failed' && 'border-danger/50 bg-danger/10 text-danger',
                  )}
                  style={
                    state === 'done'
                      ? { background: meta.color, color: '#04120c' }
                      : state === 'current'
                        ? {
                            background: `${meta.color}22`,
                            color: meta.color,
                            borderColor: `${meta.color}66`,
                            boxShadow: `0 0 18px ${meta.color}55`,
                          }
                        : undefined
                  }
                >
                  {state === 'done' ? (
                    <Check size={15} strokeWidth={3} />
                  ) : state === 'current' ? (
                    <Loader2 size={15} className="animate-spin" />
                  ) : state === 'failed' ? (
                    <AlertOctagon size={15} />
                  ) : (
                    <span className="text-[0.6rem] font-bold" style={{ color: meta.color }}>
                      {i + 1}
                    </span>
                  )}
                  {state === 'current' && (
                    <span
                      className="absolute inset-0 -z-10 animate-ping rounded-full opacity-30"
                      style={{ background: meta.color }}
                    />
                  )}
                </div>
                <span
                  className={clsx(
                    'whitespace-nowrap text-[0.65rem] font-semibold',
                    state === 'next' ? 'text-ink-faint' : 'text-ink-dim',
                  )}
                  style={state === 'done' || state === 'current' ? { color: meta.color } : undefined}
                >
                  {meta.label}
                </span>
                {!isLast && (
                  <span
                    className={clsx(
                      'absolute top-4 hidden h-0.5 w-[calc(100%+2.5rem)] left-1/2 sm:block',
                      state === 'done' ? 'bg-accent/40' : 'bg-edge',
                    )}
                  />
                )}
              </div>
            </div>
            {!isLast && (
              <div className="mb-5 hidden h-0.5 flex-1 sm:block">
                <div
                  className={clsx(
                    'h-full w-full rounded-full transition-colors',
                    state === 'done' ? 'bg-accent/50' : 'bg-edge',
                  )}
                />
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

export function FailedMarker({ stage }: { stage: string }) {
  if (stage !== 'failed') return null;
  return (
    <div className="mt-3 flex items-center gap-2 rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-xs text-danger">
      <AlertOctagon size={14} />
      This issue is in a failed state — check the runs below for the error, then use
      <code className="mono">/sloper retry</code> on the GitHub issue to recover.
    </div>
  );
}