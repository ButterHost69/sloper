import type { ReactNode } from 'react';

interface PageHeaderProps {
  eyebrow: string;
  title: string;
  subtitle?: ReactNode;
  actions?: ReactNode;
}

export function PageHeader({ eyebrow, title, subtitle, actions }: PageHeaderProps) {
  return (
    <header className="flex flex-col gap-4 border-b border-edge/70 pb-5 lg:flex-row lg:items-end lg:justify-between">
      <div className="min-w-0">
        <p className="flex items-center gap-2 text-[0.65rem] font-semibold uppercase tracking-widest text-ink-faint">
          <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true" />
          <span>Sloper console</span>
          <span className="text-edge-2" aria-hidden="true">
            /
          </span>
          <span className="text-ink-dim">{eyebrow}</span>
        </p>
        <h1 className="mt-1.5 text-2xl font-semibold tracking-tight text-ink">{title}</h1>
        {subtitle && <div className="mt-1.5 text-xs leading-5 text-ink-dim">{subtitle}</div>}
      </div>
      {actions && (
        <div className="flex w-full shrink-0 flex-wrap items-center gap-2 lg:w-auto lg:justify-end">
          {actions}
        </div>
      )}
    </header>
  );
}
