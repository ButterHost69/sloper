'use client';

import clsx from 'clsx';
import { stageMeta } from '@/lib/stages';
import { Badge } from '@/components/ui';

export function StageBadge({ stage, className }: { stage: string; className?: string }) {
  const meta = stageMeta(stage);
  return (
    <Badge color={meta.color} dot className={className}>
      {meta.label}
    </Badge>
  );
}

export function RunStatusBadge({ status }: { status: string }) {
  const colorMap: Record<string, string> = {
    running: '#38bdf8',
    completed: '#34d399',
    failed: '#f87171',
    interrupted: '#fbbf24',
  };
  return (
    <Badge color={colorMap[status] ?? '#94a3b8'} dot>
      {status}
    </Badge>
  );
}

export function RunStageBadge({ stage }: { stage: string }) {
  const colorMap: Record<string, string> = {
    spec: '#38bdf8',
    work: '#a78bfa',
    review: '#e879f9',
    fix: '#fbbf24',
    merge: '#34d399',
  };
  return (
    <Badge color={colorMap[stage] ?? '#94a3b8'} className="uppercase">
      {stage}
    </Badge>
  );
}

export function ReviewStateBadge({ state }: { state: string }) {
  const colorMap: Record<string, string> = {
    none: '#94a3b8',
    pending: '#38bdf8',
    approved: '#34d399',
    changes_requested: '#f87171',
  };
  const display = state && state.trim() ? state.replace(/_/g, ' ') : 'none';
  return (
    <Badge color={colorMap[display] ?? '#94a3b8'} dot>
      {display}
    </Badge>
  );
}

export function LabelChips({ labels }: { labels: string[] }) {
  if (!labels?.length) return <span className="text-xs text-ink-faint">—</span>;
  return (
    <span className="flex flex-wrap gap-1">
      {labels.map((l) => (
        <span
          key={l}
          className={clsx(
            'chip !text-[0.62rem]',
            l === 'triaged' && '!border-emerald/30 !text-emerald',
          )}
        >
          {l}
        </span>
      ))}
    </span>
  );
}