'use client';

import clsx from 'clsx';
import type { ReactNode } from 'react';
import { AlertTriangle, CircleAlert, RefreshCw, Loader2 } from 'lucide-react';

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={clsx('animate-spin', className)} size={16} />;
}

export function PulseDot({
  color = '#34d399',
  className,
}: {
  color?: string;
  className?: string;
}) {
  return (
    <span
      className={clsx('inline-block h-2 w-2 rounded-full', className)}
      style={{ background: color, boxShadow: `0 0 8px ${color}` }}
    />
  );
}

export function StatCard({
  label,
  value,
  sub,
  icon,
  accent = '#34d399',
  loading,
  onClick,
}: {
  label: string;
  value: ReactNode;
  sub?: ReactNode;
  icon?: ReactNode;
  accent?: string;
  loading?: boolean;
  onClick?: () => void;
}) {
  return (
    <div
      className={clsx('panel panel-hover relative overflow-hidden p-4', onClick && 'cursor-pointer')}
      onClick={onClick}
    >
      <div
        className="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full opacity-[0.12] blur-2xl"
        style={{ background: accent }}
      />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[0.7rem] font-semibold uppercase tracking-wider text-ink-faint">
            {label}
          </p>
          {loading ? (
            <div className="skeleton mt-2 h-7 w-20" />
          ) : (
            <p className="mt-1 text-2xl font-semibold tabular-nums text-ink">{value}</p>
          )}
          {sub && <div className="mt-1 text-xs text-ink-dim">{sub}</div>}
        </div>
        {icon && (
          <div
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-edge-2"
            style={{ color: accent, background: 'rgba(255,255,255,0.02)' }}
          >
            {icon}
          </div>
        )}
      </div>
    </div>
  );
}

export function Badge({
  children,
  color = '#94a3b8',
  className,
  dot,
}: {
  children: ReactNode;
  color?: string;
  className?: string;
  dot?: boolean;
}) {
  return (
    <span
      className={clsx('chip', className)}
      style={{ color, borderColor: `${color}44`, background: `${color}14` }}
    >
      {dot && <PulseDot color={color} className="!h-1.5 !w-1.5" />}
      {children}
    </span>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={clsx('skeleton', className)} />;
}

export function EmptyState({
  title,
  hint,
  icon,
}: {
  title: string;
  hint?: string;
  icon?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 py-14 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-full border border-edge-2 text-ink-faint">
        {icon ?? <CircleAlert size={20} />}
      </div>
      <p className="text-sm font-medium text-ink-dim">{title}</p>
      {hint && <p className="max-w-xs text-xs text-ink-faint">{hint}</p>}
    </div>
  );
}

export function ErrorState({
  error,
  onRetry,
  compact,
}: {
  error: Error | null;
  onRetry?: () => void;
  compact?: boolean;
}) {
  return (
    <div
      className={clsx(
        'flex flex-col items-center justify-center gap-3 text-center',
        compact ? 'py-8' : 'py-16',
      )}
    >
      <div className="flex h-11 w-11 items-center justify-center rounded-full border border-danger/30 bg-danger/10 text-danger">
        <AlertTriangle size={18} />
      </div>
      <div>
        <p className="text-sm font-semibold text-ink">Could not reach this sloper instance</p>
        <p className="mt-1 max-w-md text-xs text-ink-faint">
          {error?.message ?? 'Unknown error'}
        </p>
      </div>
      {onRetry && (
        <button className="btn btn-ghost" onClick={onRetry}>
          <RefreshCw size={14} /> Retry
        </button>
      )}
    </div>
  );
}

export function SectionHeader({
  title,
  sub,
  action,
}: {
  title: string;
  sub?: string;
  action?: ReactNode;
}) {
  return (
    <div className="mb-3 flex items-end justify-between gap-3">
      <div>
        <h2 className="text-sm font-semibold text-ink">{title}</h2>
        {sub && <p className="mt-0.5 text-xs text-ink-faint">{sub}</p>}
      </div>
      {action}
    </div>
  );
}

export function Panel({
  children,
  className,
  title,
  action,
  bodyClassName,
}: {
  children: ReactNode;
  className?: string;
  title?: string;
  action?: ReactNode;
  bodyClassName?: string;
}) {
  return (
    <section className={clsx('panel', className)}>
      {(title || action) && (
        <header className="flex items-center justify-between border-b border-edge px-4 py-3">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-ink-dim">
            {title}
          </h3>
          {action}
        </header>
      )}
      <div className={clsx('p-4', bodyClassName)}>{children}</div>
    </section>
  );
}

export function ProgressBar({
  value,
  max,
  color = '#34d399',
  className,
}: {
  value: number;
  max: number;
  color?: string;
  className?: string;
}) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;
  return (
    <div className={clsx('h-1.5 w-full overflow-hidden rounded-full bg-edge', className)}>
      <div
        className="h-full rounded-full transition-all duration-500"
        style={{ width: `${pct}%`, background: color }}
      />
    </div>
  );
}