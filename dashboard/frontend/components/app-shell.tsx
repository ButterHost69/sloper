'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import clsx from 'clsx';
import {
  Activity,
  Check,
  ChevronsUpDown,
  ChevronsLeft,
  ChevronsRight,
  GitPullRequest,
  LayoutDashboard,
  ListChecks,
  Menu,
  Moon,
  Plus,
  RefreshCw,
  ScrollText,
  Server,
  Settings,
  Sparkles,
  Sun,
  X,
} from 'lucide-react';
import { RavenLogo } from '@/components/raven-logo';
import { useInstances } from '@/components/instance-context';
import { useHealth } from '@/components/health-context';
import { safeHost } from '@/lib/instances';

type Theme = 'dark' | 'light';

function readStorage(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

function writeStorage(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // Storage can be unavailable in partitioned or privacy-restricted contexts.
  }
}

const NAV_GROUPS: Array<{
  label: string;
  items: Array<{
    href: string;
    label: string;
    icon: typeof LayoutDashboard;
    planned?: boolean;
  }>;
}> = [
  {
    label: 'Monitor',
    items: [
      { href: '/', label: 'Overview', icon: LayoutDashboard },
      { href: '/issues', label: 'Issues', icon: ListChecks },
      { href: '/pulls', label: 'Pull requests', icon: GitPullRequest },
      { href: '/runs', label: 'Runs', icon: Activity },
      { href: '/events', label: 'Events', icon: ScrollText },
    ],
  },
  {
    label: 'Workspace',
    items: [
      { href: '/instances', label: 'Instances', icon: Server },
      { href: '/roadmap', label: 'Roadmap', icon: Sparkles, planned: true },
    ],
  },
];

function isActivePath(pathname: string, href: string): boolean {
  if (href === '/') return pathname === '/';
  return pathname === href || pathname.startsWith(`${href}/`);
}

function pageMeta(pathname: string): { section: string; title: string } {
  if (/^\/issues\/\d+/.test(pathname)) return { section: 'Issues', title: 'Issue detail' };
  if (pathname.startsWith('/issues')) return { section: 'Monitor', title: 'Issues' };
  if (pathname.startsWith('/pulls')) return { section: 'Monitor', title: 'Pull requests' };
  if (pathname.startsWith('/runs')) return { section: 'Monitor', title: 'Runs' };
  if (pathname.startsWith('/events')) return { section: 'Monitor', title: 'Events' };
  if (pathname.startsWith('/instances')) return { section: 'Workspace', title: 'Instances' };
  if (pathname.startsWith('/roadmap')) return { section: 'Workspace', title: 'Roadmap' };
  return { section: 'Monitor', title: 'Overview' };
}

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const meta = pageMeta(pathname);
  const { instances, active, setActive } = useInstances();
  const { health, error: healthError, refresh, refreshing } = useHealth();
  const [pickerOpen, setPickerOpen] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const [isDesktop, setIsDesktop] = useState(true);
  const [theme, setTheme] = useState<Theme>('light');
  const pickerRef = useRef<HTMLDivElement>(null);
  const pickerButtonRef = useRef<HTMLButtonElement>(null);
  const mobileCloseRef = useRef<HTMLButtonElement>(null);
  const mobileMenuButtonRef = useRef<HTMLButtonElement>(null);

  const connected = health?.status === 'ok' && !healthError;
  const rail = isDesktop && collapsed;

  useEffect(() => {
    const stored = readStorage('sloper-theme');
    const initial: Theme = stored === 'light' || stored === 'dark' ? stored : 'light';
    setTheme(initial);
    document.documentElement.dataset.theme = initial;
  }, []);

  useEffect(() => {
    setCollapsed(readStorage('sloper-sidebar-collapsed') === 'true');
  }, []);

  useEffect(() => {
    const media = window.matchMedia('(min-width: 1024px)');
    const update = () => {
      setIsDesktop(media.matches);
      setMobileOpen(false);
    };
    update();
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, []);

  useEffect(() => {
    if (!pickerOpen) return;
    const close = (event: PointerEvent) => {
      if (!pickerRef.current?.contains(event.target as Node)) setPickerOpen(false);
    };
    const escape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setPickerOpen(false);
    };
    window.addEventListener('pointerdown', close);
    window.addEventListener('keydown', escape);
    return () => {
      window.removeEventListener('pointerdown', close);
      window.removeEventListener('keydown', escape);
      pickerButtonRef.current?.focus();
    };
  }, [pickerOpen]);

  useEffect(() => {
    if (!mobileOpen) return;
    mobileCloseRef.current?.focus();
    const escape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setMobileOpen(false);
    };
    window.addEventListener('keydown', escape);
    return () => {
      window.removeEventListener('keydown', escape);
      if (!isDesktop && !window.matchMedia('(min-width: 1024px)').matches) {
        mobileMenuButtonRef.current?.focus();
      }
    };
  }, [isDesktop, mobileOpen]);

  useEffect(() => {
    setMobileOpen(false);
    setPickerOpen(false);
  }, [pathname]);

  const toggleTheme = () => {
    const next: Theme = theme === 'dark' ? 'light' : 'dark';
    setTheme(next);
    writeStorage('sloper-theme', next);
    document.documentElement.dataset.theme = next;
  };

  const toggleSidebar = () => {
    setCollapsed((value) => {
      writeStorage('sloper-sidebar-collapsed', String(!value));
      return !value;
    });
  };

  const handleDrawerKeyDown = (event: React.KeyboardEvent<HTMLElement>) => {
    if (isDesktop || !mobileOpen || event.key !== 'Tab') return;
    const focusable = Array.from(
      event.currentTarget.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
      ),
    ).filter((element) => element.offsetParent !== null);
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  };

  return (
    <div className="min-h-screen bg-transparent">
      {mobileOpen && (
        <button
          className="fixed inset-0 z-40 bg-black/45 backdrop-blur-[2px] lg:hidden"
          aria-label="Close navigation"
          tabIndex={-1}
          onClick={() => setMobileOpen(false)}
        />
      )}

      <aside
        aria-hidden={!isDesktop && !mobileOpen ? true : undefined}
        id="primary-navigation"
        aria-label="Primary navigation"
        inert={!isDesktop && !mobileOpen}
        onKeyDown={handleDrawerKeyDown}
        className={clsx(
          'glass-chrome glass-sidebar fixed inset-y-0 left-0 z-50 flex flex-col border-r border-edge transition-[width,transform] duration-200 ease-out',
          rail ? 'w-20' : 'w-[252px]',
          mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
        )}
      >
        <div
          className={clsx(
            'flex h-16 shrink-0 items-center border-b border-edge',
            rail ? 'gap-1 px-1' : 'gap-2 px-4',
          )}
        >
          <Link
            href="/"
            className={clsx(
              'flex min-w-0 items-center gap-2.5 rounded-lg text-ink',
              rail ? 'mx-auto' : 'mr-auto',
            )}
            aria-label="Sloper home"
          >
            <span className="flex h-8 w-8 shrink-0 items-center justify-center">
              <RavenLogo size={32} />
            </span>
            {!rail && (
              <span className="min-w-0 text-[15px] font-semibold leading-4 tracking-[-0.01em]">
                sloper
              </span>
            )}
          </Link>
          <button
            className="btn btn-ghost hidden !h-8 !min-h-8 !w-8 !p-0 lg:inline-flex"
            onClick={toggleSidebar}
            aria-label={rail ? 'Expand sidebar' : 'Collapse sidebar'}
            title={rail ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {rail ? <ChevronsRight size={15} /> : <ChevronsLeft size={15} />}
          </button>
          <button
            ref={mobileCloseRef}
            className="btn btn-ghost !h-8 !min-h-8 !w-8 !p-0 lg:hidden"
            onClick={() => setMobileOpen(false)}
            aria-label="Close navigation"
          >
            <X size={16} />
          </button>
        </div>

        {!rail && (
          <div className="px-3 pt-3" ref={pickerRef}>
            <button
              ref={pickerButtonRef}
              onClick={() => setPickerOpen((value) => !value)}
              className={clsx(
                'glass-chrome glass-control flex w-full items-center gap-2.5 rounded-xl border px-3 py-2.5 text-left transition hover:border-edge-2',
                pickerOpen && 'border-accent-dim/60',
              )}
              aria-expanded={pickerOpen}
              aria-controls="instance-picker-menu"
            >
              <span
                className={clsx(
                  'h-2 w-2 shrink-0 rounded-full',
                  connected ? 'bg-emerald shadow-[0_0_9px_rgba(78,209,126,0.6)]' : 'bg-danger',
                )}
              />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-[13px] font-medium text-ink">
                  {active?.name ?? 'Choose an instance'}
                </span>
                <span className="mt-0.5 block truncate text-[10px] text-ink-faint">
                  {active ? safeHost(active.url) || 'unknown host' : 'Read-only observability'}
                </span>
              </span>
              <ChevronsUpDown size={14} className="text-ink-faint" />
            </button>

            {pickerOpen && (
              <div
                id="instance-picker-menu"
                aria-label="Choose an instance"
                className="glass-chrome glass-popover absolute left-3 right-3 top-[70px] z-50 overflow-hidden rounded-2xl border border-edge-2 p-1.5"
              >
                {instances.length === 0 && (
                  <p className="px-3 py-3 text-xs leading-5 text-ink-faint">
                    No instances configured yet.
                  </p>
                )}
                {instances.map((instance) => (
                  <button
                    key={instance.id}
                    onClick={() => {
                      setActive(instance.id);
                      setPickerOpen(false);
                    }}
                    className={clsx(
                      'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left transition hover:bg-panel-2',
                      instance.id === active?.id && 'bg-panel-2',
                    )}
                    aria-pressed={instance.id === active?.id}
                  >
                    <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-accent" />
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-xs font-medium text-ink">{instance.name}</span>
                      <span className="block truncate text-[10px] text-ink-faint">{instance.url}</span>
                    </span>
                    {instance.id === active?.id && <Check size={14} className="text-accent" />}
                  </button>
                ))}
                <Link
                  href="/instances"
                  onClick={() => setPickerOpen(false)}
                  className="mt-1 flex items-center gap-2 border-t border-edge px-3 py-2.5 text-xs font-medium text-accent transition hover:bg-panel-2"
                >
                  <Plus size={13} /> Manage instances
                </Link>
              </div>
            )}
          </div>
        )}

        {rail && (
          <div className="flex justify-center px-3 pt-3">
            <Link
              href="/instances"
              className={clsx(
                'flex h-9 w-9 items-center justify-center rounded-xl border transition',
                connected
                  ? 'border-emerald/25 bg-emerald/10 text-emerald'
                  : 'border-edge bg-panel-2 text-ink-faint hover:text-ink',
              )}
              title={active?.name ?? 'No active instance'}
              aria-label={active?.name ?? 'No active instance'}
            >
              <span className="h-2 w-2 rounded-full bg-current" />
            </Link>
          </div>
        )}

        <nav className="min-h-0 flex-1 overflow-y-auto px-3 py-4" aria-label="Primary navigation">
          {NAV_GROUPS.map((group) => (
            <div key={group.label} className="mb-5 last:mb-0">
              {!rail && (
                <p className="mb-1.5 px-3 text-[10px] font-semibold uppercase tracking-[0.13em] text-ink-faint">
                  {group.label}
                </p>
              )}
              <div className="space-y-0.5">
                {group.items.map(({ href, label, icon: Icon, planned }) => {
                  const activePage = isActivePath(pathname, href);
                  return (
                    <Link
                      key={href}
                      href={href}
                      onClick={() => setMobileOpen(false)}
                      className={clsx(
                        'group relative flex h-9 items-center rounded-lg text-[13px] font-medium transition',
                        rail ? 'justify-center px-0' : 'gap-2.5 px-3',
                        activePage
                          ? 'bg-accent/10 text-ink'
                          : 'text-ink-dim hover:bg-panel-2 hover:text-ink',
                      )}
                      aria-current={activePage ? 'page' : undefined}
                      aria-label={rail ? label : undefined}
                      title={rail ? label : undefined}
                    >
                      {activePage && (
                        <span className="absolute inset-y-2 left-0 w-0.5 rounded-full bg-accent" />
                      )}
                      <Icon
                        size={16}
                        className={clsx(
                          'shrink-0 transition-colors',
                          activePage ? 'text-accent' : 'text-ink-faint group-hover:text-ink-dim',
                        )}
                      />
                      {!rail && <span className="truncate">{label}</span>}
                      {!rail && planned && (
                        <span className="ml-auto rounded-full border border-edge-2 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wide text-ink-faint">
                          Preview
                        </span>
                      )}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="shrink-0 border-t border-edge p-3">
          {rail ? (
            <Link
              href="/instances"
              className="flex h-9 items-center justify-center rounded-lg text-ink-faint transition hover:bg-panel-2 hover:text-ink"
              aria-label="Instance settings"
              title="Instance settings"
            >
              <Settings size={16} />
            </Link>
          ) : (
            <div className="glass-chrome glass-control rounded-xl border p-3">
              <div className="flex items-center gap-2">
                <span
                  className={clsx(
                    'h-2 w-2 shrink-0 rounded-full',
                    connected ? 'live-dot' : 'bg-danger',
                  )}
                />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-xs font-medium text-ink">
                    {health?.repo ?? active?.name ?? 'No instance selected'}
                  </p>
                  <p className="mt-0.5 text-[10px] text-ink-faint">
                    {connected ? 'API connected' : 'API unavailable'}
                  </p>
                </div>
                <Link
                  href="/instances"
                  className="text-ink-faint transition hover:text-ink"
                  aria-label="Instance settings"
                >
                  <Settings size={14} />
                </Link>
              </div>
            </div>
          )}
        </div>
      </aside>

      <div
        className={clsx(
          'min-w-0 transition-[margin] duration-200 ease-out',
          rail ? 'lg:ml-20' : 'lg:ml-[252px]',
        )}
      >
        <header
          inert={!isDesktop && mobileOpen}
          className="glass-chrome glass-topbar sticky top-0 z-30 flex h-16 items-center border-b border-edge px-4 sm:px-6 lg:px-8"
        >
          <button
            ref={mobileMenuButtonRef}
            className="btn btn-ghost mr-2 !h-9 !min-h-9 !w-9 !p-0 lg:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="Open navigation"
            aria-expanded={mobileOpen}
            aria-controls="primary-navigation"
          >
            <Menu size={18} />
          </button>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 text-xs">
              <span className="hidden text-ink-faint sm:inline">{meta.section}</span>
              <span className="hidden text-ink-faint sm:inline">/</span>
              <span className="truncate font-medium text-ink">{meta.title}</span>
            </div>
          </div>

          <div className="flex items-center gap-1.5">
            <Link
              href="/instances"
              className="glass-chrome glass-control hidden h-9 items-center gap-2 rounded-full border px-3 text-xs text-ink-dim transition hover:border-edge-2 hover:text-ink sm:flex"
            >
              <span
                className={clsx(
                  'h-1.5 w-1.5 rounded-full',
                  connected ? 'bg-emerald' : 'bg-danger',
                )}
              />
              {connected ? 'Connected' : 'Offline'}
            </Link>
            <button
              className="btn btn-ghost !h-9 !min-h-9 !w-9 !p-0"
              onClick={refresh}
              disabled={refreshing}
              aria-label="Refresh instance health"
              title="Refresh instance health"
            >
              <RefreshCw size={15} className={refreshing ? 'animate-spin' : ''} />
            </button>
            <button
              className="btn btn-ghost !h-9 !min-h-9 !w-9 !p-0"
              onClick={toggleTheme}
              aria-label={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
            >
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
            </button>
          </div>
        </header>

        <main inert={!isDesktop && mobileOpen} className="px-4 py-5 sm:px-6 sm:py-7 lg:px-8 lg:py-8">
          <div className="mx-auto max-w-[1480px] page-enter">{children}</div>
        </main>
      </div>
    </div>
  );
}
