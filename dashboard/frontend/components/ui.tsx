'use client';

import clsx from 'clsx';
import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { AlertTriangle, CircleAlert, Loader2, RefreshCw } from 'lucide-react';

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={clsx('animate-spin', className)} size={16} aria-hidden="true" />;
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
      className={clsx('inline-block h-2 w-2 shrink-0 rounded-full', className)}
      style={{ background: color, boxShadow: `0 0 8px ${color}` }}
      aria-hidden="true"
    />
  );
}

export function StatCard({
  label,
  value,
  sub,
  icon,
  accent = '#7aaaff',
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
  const content = (
    <>
      <div
        className="pointer-events-none absolute -right-8 -top-10 h-28 w-28 rounded-full opacity-[0.09] blur-3xl"
        style={{ background: accent }}
      />
      <div className="relative flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[10px] font-semibold uppercase tracking-[0.1em] text-ink-faint">{label}</p>
          {loading ? (
            <div className="skeleton mt-2.5 h-7 w-20" />
          ) : (
            <p className="mt-1.5 text-[26px] font-semibold leading-none tracking-[-0.035em] tabular-nums text-ink">
              {value}
            </p>
          )}
          {sub && <div className="mt-2 text-[11px] leading-4 text-ink-dim">{sub}</div>}
        </div>
        {icon && (
          <div
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-[9px] border border-edge bg-panel-2"
            style={{ color: accent }}
          >
            {icon}
          </div>
        )}
      </div>
    </>
  );

  const className = clsx(
    'panel panel-hover relative min-h-[112px] overflow-hidden p-4 text-left',
    onClick && 'hover:cursor-pointer',
  );

  if (onClick) {
    return (
      <button type="button" className={className} onClick={onClick}>
        {content}
      </button>
    );
  }

  return <div className={className}>{content}</div>;
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
      style={{ color, borderColor: `${color}55`, background: `${color}16` }}
    >
      {dot && <PulseDot color={color} className="!h-1.5 !w-1.5" />}
      {children}
    </span>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={clsx('skeleton', className)} aria-hidden="true" />;
}

export function EmptyState({
  title,
  hint,
  icon,
  action,
}: {
  title: string;
  hint?: string;
  icon?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 px-5 py-14 text-center">
      <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-edge bg-panel-2 text-ink-faint">
        {icon ?? <CircleAlert size={19} aria-hidden="true" />}
      </div>
      <div>
        <p className="text-sm font-medium text-ink">{title}</p>
        {hint && <p className="mx-auto mt-1.5 max-w-sm text-xs leading-5 text-ink-faint">{hint}</p>}
      </div>
      {action}
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
        'flex flex-col items-center justify-center gap-3 px-5 text-center',
        compact ? 'py-8' : 'py-14',
      )}
      role="alert"
    >
      <div className="flex h-11 w-11 items-center justify-center rounded-xl border border-danger/25 bg-danger/10 text-danger">
        <AlertTriangle size={18} aria-hidden="true" />
      </div>
      <div>
        <p className="text-sm font-semibold text-ink">Could not reach this Sloper instance</p>
        <p className="mx-auto mt-1.5 max-w-md break-words text-xs leading-5 text-ink-faint">
          {error?.message ?? 'Unknown error'}
        </p>
      </div>
      {onRetry && (
        <button className="btn btn-ghost" onClick={onRetry} type="button">
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
        <h2 className="text-sm font-semibold tracking-[-0.01em] text-ink">{title}</h2>
        {sub && <p className="mt-1 text-[11px] leading-4 text-ink-faint">{sub}</p>}
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
    <section className={clsx('panel overflow-hidden', className)}>
      {(title || action) && (
        <header className="flex min-h-12 items-center justify-between gap-3 border-b border-edge px-4 py-3 sm:px-5">
          <h3 className="text-xs font-semibold tracking-[-0.005em] text-ink-dim">{title}</h3>
          {action}
        </header>
      )}
      <div className={clsx('p-4 sm:p-5', bodyClassName)}>{children}</div>
    </section>
  );
}

export function ProgressBar({
  value,
  max,
  color = '#7aaaff',
  className,
  label,
}: {
  value: number;
  max: number;
  color?: string;
  className?: string;
  label?: string;
}) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;
  return (
    <div
      className={clsx('h-1.5 w-full overflow-hidden rounded-full bg-edge', className)}
      role="progressbar"
      aria-label={label}
      aria-valuemin={0}
      aria-valuemax={max}
      aria-valuenow={value}
    >
      <div
        className="h-full rounded-full transition-all duration-500"
        style={{ width: `${pct}%`, background: color }}
      />
    </div>
  );
}

export function IconButton({
  className,
  children,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button type="button" className={clsx('btn btn-ghost !p-2', className)} {...props}>
      {children}
    </button>
  );
}
