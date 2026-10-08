'use client';

import { Fragment } from 'react';
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
    <div className="overflow-x-auto pb-1">
      <div className="flex px-1">
        {PIPELINE_ORDER.map((key, i) => {
          const meta = stageMeta(key);
          const state = states[key] ?? 'next';
          const isLast = i === PIPELINE_ORDER.length - 1;
          return (
            <Fragment key={key}>
              <div className="relative flex w-9 shrink-0 flex-col items-center">
              {/* node */}
              <div
                className={clsx(
                  'relative flex h-9 w-9 items-center justify-center rounded-full border transition-all',
                  state === 'done' && 'border-transparent',
                  state === 'current' && 'border-accent/50 bg-accent/10 text-accent shadow-[0_0_16px_rgba(65,118,230,0.22)]',
                  state === 'next' && 'border-edge-2',
                  state === 'failed' && 'border-danger/50 bg-danger/10 text-danger',
                )}
                style={state === 'done' ? { background: meta.color, color: '#04120c' } : undefined}
              >
                {state === 'done' ? (
                  <Check size={15} strokeWidth={3} />
                ) : state === 'current' ? (
                  <Loader2 size={15} className="animate-spin" />
                ) : state === 'failed' ? (
                  <AlertOctagon size={15} />
                ) : (
                  <span className="text-[0.6rem] font-bold text-ink-faint">{i + 1}</span>
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
                  'mt-1.5 whitespace-nowrap text-[0.65rem] font-semibold',
                  state === 'next' ? 'text-ink-faint' : state === 'current' ? 'text-accent' : 'text-ink',
                )}
              >
                {meta.label}
              </span>
            </div>
            {!isLast && (
              // Single connector per gap, aligned to the dot's vertical center
              // (dot is h-9 → center at 18px; line is 2px → top at 17px).
              <div className="mt-[17px] hidden h-0.5 min-w-4 flex-1 self-start sm:block">
                <div
                  className={clsx(
                    'h-full w-full rounded-full transition-colors',
                    state === 'done' ? 'bg-accent/40' : 'bg-edge',
                  )}
                />
              </div>
            )}
            </Fragment>
          );
        })}
      </div>
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