'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useState } from 'react';
import clsx from 'clsx';
import {
  Activity,
  CheckCircle2,
  ChevronsUpDown,
  GitPullRequest,
  LayoutDashboard,
  ListChecks,
  Plus,
  Settings,
} from 'lucide-react';
import { useInstances } from '@/components/instance-context';
import { useHealth } from '@/components/health-context';
import { PulseDot } from '@/components/ui';
import { safeHost } from '@/lib/instances';

const NAV = [
  { href: '/', label: 'Overview', icon: LayoutDashboard },
  { href: '/issues', label: 'Issues', icon: ListChecks },
  { href: '/pulls', label: 'Pull Requests', icon: GitPullRequest },
  { href: '/runs', label: 'Runs', icon: Activity },
  { href: '/instances', label: 'Instances', icon: Settings },
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { instances, active, setActive } = useInstances();
  const [pickerOpen, setPickerOpen] = useState(false);
  const { health } = useHealth();

  const connected = !!health && health.status === 'ok';

  return (
    <div className="flex min-h-screen">
      {/* ─── Sidebar ─────────────────────────────────────────────── */}
      <aside className="fixed inset-y-0 left-0 z-40 flex w-60 flex-col border-r border-edge bg-panel">
        <Link href="/" className="flex items-center gap-2.5 px-5 py-5">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-dim text-black shadow-[0_0_20px_rgba(52,211,153,0.35)]">
            <Activity size={17} strokeWidth={2.5} />
          </span>
          <span className="text-lg font-bold tracking-tight text-ink">
            sloper
            <span className="ml-1.5 rounded border border-edge-2 px-1 py-px align-middle text-[0.55rem] font-semibold uppercase tracking-widest text-ink-faint">
              console
            </span>
          </span>
        </Link>

        {/* Instance picker */}
        <div className="px-3">
          <button
            onClick={() => setPickerOpen((v) => !v)}
            className={clsx(
              'flex w-full items-center gap-2 rounded-lg border border-edge-2 bg-panel-2 px-3 py-2 text-left transition hover:border-edge-2/70',
              pickerOpen && 'border-accent-dim/60',
            )}
          >
            {connected ? (
              <PulseDot color="#34d399" />
            ) : (
              <PulseDot color="#f87171" />
            )}
            <span className="min-w-0 flex-1">
              <span className="block truncate text-sm font-medium text-ink">
                {active?.name ?? 'No instance'}
              </span>
              <span className="block truncate text-[0.65rem] text-ink-faint">
                {active ? safeHost(active.url) || 'unknown host' : 'Add an instance'}
              </span>
            </span>
            <ChevronsUpDown size={14} className="text-ink-faint" />
          </button>

          {pickerOpen && (
            <div className="mt-2 overflow-hidden rounded-lg border border-edge-2 bg-panel-2 shadow-xl shadow-black/50">
              {instances.length === 0 && (
                <p className="px-3 py-3 text-xs text-ink-faint">No instances configured.</p>
              )}
              {instances.map((inst) => (
                <button
                  key={inst.id}
                  onClick={() => {
                    setActive(inst.id);
                    setPickerOpen(false);
                  }}
                  className={clsx(
                    'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition hover:bg-white/5',
                    inst.id === active?.id && 'bg-white/5 text-ink',
                  )}
                >
                  <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-accent" />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate">{inst.name}</span>
                    <span className="block truncate text-[0.65rem] text-ink-faint">
                      {inst.url}
                    </span>
                  </span>
                  {inst.id === active?.id && <CheckCircle2 size={14} className="text-accent" />}
                </button>
              ))}
              <Link
                href="/instances"
                onClick={() => setPickerOpen(false)}
                className="flex w-full items-center gap-2 border-t border-edge px-3 py-2 text-sm text-accent hover:bg-white/5"
              >
                <Plus size={14} /> Manage instances
              </Link>
            </div>
          )}
        </div>

        <nav className="mt-4 flex-1 space-y-1 px-3">
          {NAV.map(({ href, label, icon: Icon }) => {
            const activePage = pathname === href;
            return (
              <Link
                key={href}
                href={href}
                className={clsx(
                  'group flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition',
                  activePage
                    ? 'bg-white/[0.06] text-ink'
                    : 'text-ink-dim hover:bg-white/[0.03] hover:text-ink',
                )}
              >
                <Icon
                  size={16}
                  className={clsx(
                    'transition',
                    activePage ? 'text-accent' : 'text-ink-faint group-hover:text-ink-dim',
                  )}
                />
                {label}
                {activePage && <span className="ml-auto h-1.5 w-1.5 rounded-full bg-accent" />}
              </Link>
            );
          })}
        </nav>

        <div className="border-t border-edge p-3">
          <div className="flex items-center gap-2 rounded-lg border border-edge bg-panel-2 px-3 py-2">
            <PulseDot color={connected ? '#34d399' : '#f87171'} />
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-medium text-ink">
                {health?.repo ?? active?.name ?? 'Offline'}
              </p>
              <p className="truncate text-[0.65rem] text-ink-faint">
                {connected ? 'Connected' : 'Unreachable'}
              </p>
            </div>
            <Link
              href="/instances"
              className="text-ink-faint transition hover:text-ink"
              title="Instance settings"
            >
              <Settings size={14} />
            </Link>
          </div>
        </div>
      </aside>

      {/* ─── Main content ────────────────────────────────────────── */}
      <main className="ml-60 min-w-0 flex-1">
        <div className="mx-auto max-w-7xl px-6 py-6 page-enter">{children}</div>
      </main>
    </div>
  );
}