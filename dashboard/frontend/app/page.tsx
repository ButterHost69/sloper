'use client';

import { useMemo } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Activity,
  ArrowRight,
  Bot,
  Boxes,
  CircleDot,
  Clock3,
  CodeXml,
  Database,
  ExternalLink,
  GitMerge,
  GitPullRequest,
  ListChecks,
  RefreshCw,
  Server,
  Sparkles,
  TriangleAlert,
  Workflow,
} from 'lucide-react';
import { useInstances } from '@/components/instance-context';
import { useHealth } from '@/components/health-context';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { formatBytes, timeAgo } from '@/lib/format';
import type { EventRecord } from '@/lib/types';
import {
  Badge,
  EmptyState,
  ErrorState,
  Panel,
  PulseDot,
  Skeleton,
  StaleDataNotice,
  StatCard,
} from '@/components/ui';
import { PipelineFunnel } from '@/components/pipeline-funnel';
import { FailuresPanel } from '@/components/failures-panel';
import { RunsView } from '@/components/runs-view';
import { EventFeed } from '@/components/event-feed';

const PLANNED = [
  {
    icon: Workflow,
    title: 'Conditional loops',
    description: 'Compose Spec → Work → Review → Fix → Merge with reusable exit conditions.',
  },
  {
    icon: Boxes,
    title: 'Runner pools',
    description: 'Route work to local, containerized, or remote workers with health-aware placement.',
  },
  {
    icon: CodeXml,
    title: 'Extension packs',
    description: 'Discover and share agent recipes, project context, and pipeline templates.',
  },
];

function EmptyConnection() {
  return (
    <Panel className="surface-grid overflow-hidden" bodyClassName="p-0">
      <div className="grid gap-8 p-6 sm:p-8 lg:grid-cols-[1fr_0.9fr] lg:items-center">
        <div>
          <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-xl border border-accent/20 bg-accent/10 text-accent">
            <Server size={20} />
          </div>
          <p className="text-xs font-medium uppercase tracking-[0.12em] text-accent">Connect your runtime</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em] text-ink">
            Sloper is ready for an instance
          </h2>
          <p className="mt-3 max-w-xl text-sm leading-6 text-ink-dim">
            Point this console at a running <code className="mono text-accent">sloper-web</code> process.
            The browser reads its existing SQLite-backed API directly; no state is written or mutated.
          </p>
          <Link href="/instances" className="btn btn-primary mt-5">
            <Server size={14} /> Configure an instance
          </Link>
        </div>
        <ol className="space-y-3">
          {[
            ['Build the API', 'Run make build-web on the machine hosting Sloper.'],
            ['Expose the port', 'The API listens on 127.0.0.1:8080 by default.'],
            ['Add its URL', 'Use a bearer token when SLOPER_WEB_TOKEN is enabled.'],
          ].map(([title, description], index) => (
            <li key={title} className="flex gap-3 rounded-xl border border-edge bg-panel/90 p-3.5">
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-edge-2 bg-panel-2 text-[11px] font-semibold text-ink-dim">
                {index + 1}
              </span>
              <div>
                <p className="text-xs font-semibold text-ink">{title}</p>
                <p className="mt-1 text-[11px] leading-4 text-ink-faint">{description}</p>
              </div>
            </li>
          ))}
        </ol>
      </div>
    </Panel>
  );
}

export default function OverviewPage() {
  const { active } = useInstances();
  const { health, error: healthError } = useHealth();
  const router = useRouter();

  const { data: summary, loading: summaryLoading, error, refresh, refreshing } = useInstanceData(
    (base) => api.summary(base),
    10000,
  );
  const {
    data: events,
    error: eventsError,
    refresh: refreshEvents,
    refreshing: eventsRefreshing,
  } = useInstanceData((base) => api.events(base, 5000), 10000);
  const {
    data: runs,
    error: runsError,
    refresh: refreshRuns,
    refreshing: runsRefreshing,
  } = useInstanceData((base) => api.runs(base, 300), 15000);
  const {
    data: repo,
    error: repoError,
    refresh: refreshRepo,
    refreshing: repoRefreshing,
  } = useInstanceData((base) => api.repo(base), 60000);

  const lastTick = useMemo(() => {
    const list = events?.events ?? [];
    const startedIndex = list.findIndex((event) => event.event_type === 'tick.started');
    const started = startedIndex >= 0 ? list[startedIndex] : undefined;
    const completed = startedIndex >= 0
      ? list.slice(0, startedIndex).find((event) => event.event_type === 'tick.completed')
      : undefined;
    let durationS: number | null = null;
    if (started && completed) {
      const start = new Date(started.created_at).getTime();
      const end = new Date(completed.created_at).getTime();
      if (!Number.isNaN(start) && !Number.isNaN(end) && end >= start) {
        durationS = Math.round((end - start) / 1000);
      }
    }
    return { started, durationS };
  }, [events]);

  const runList = runs?.runs ?? [];
  const recentRuns = runList.slice(0, 7);
  const tickCount = summary?.events?.by_type?.['tick.completed'] ?? 0;
  const connected = health?.status === 'ok' && !healthError;
  const secondaryError = eventsError ?? runsError ?? repoError;
  const refreshingAll =
    refreshing || eventsRefreshing || runsRefreshing || repoRefreshing || summaryLoading;

  const refreshAll = () => {
    refresh();
    refreshEvents();
    refreshRuns();
    refreshRepo();
  };

  return (
    <div className="space-y-6">
      <section className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.14em] text-accent">
            <Activity size={12} /> Live operations
          </div>
          <h1 className="mt-2 truncate text-[28px] font-semibold leading-tight tracking-[-0.04em] text-ink sm:text-[32px]">
            {summary?.repo ?? repo?.name ?? 'Mission control'}
          </h1>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-ink-dim">
            Follow issues from first signal to merged pull request, with every agent run and decision in view.
          </p>
        </div>
        <button
          type="button"
          className="btn self-start sm:self-auto"
          onClick={refreshAll}
          disabled={!active || refreshingAll}
        >
          <RefreshCw size={14} className={refreshingAll ? 'animate-spin' : ''} />
          Refresh dashboard
        </button>
      </section>

      {!active ? (
        <EmptyConnection />
      ) : error && !summary ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : (
        <>
          {error && summary && <StaleDataNotice error={error} onRetry={refresh} label="dashboard" />}
          {secondaryError && (events || runs || repo) && (
            <StaleDataNotice error={secondaryError} onRetry={refreshAll} label="dashboard data" />
          )}
          <section className="panel surface-grid relative overflow-hidden p-4 sm:p-5">
            <div className="pointer-events-none absolute -right-16 -top-20 h-56 w-56 rounded-full bg-accent/10 blur-3xl" />
            <div className="relative flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div className="flex items-start gap-3">
                <span
                  className={`mt-1 h-2.5 w-2.5 shrink-0 rounded-full ${connected ? 'live-dot' : 'bg-danger'}`}
                />
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="text-sm font-semibold text-ink">
                      {connected ? 'The coordinator is observable' : 'Waiting for the instance API'}
                    </p>
                    <Badge color={connected ? '#4ed17e' : '#f87171'} dot>
                      {connected ? 'Live' : 'Offline'}
                    </Badge>
                  </div>
                  <p className="mt-1.5 text-xs leading-5 text-ink-faint">
                    {active.name} · {health?.version ? `v${health.version}` : 'read-only API'}
                    {health?.go ? ` · Go ${health.go}` : ''}
                  </p>
                </div>
              </div>

              <div className="flex flex-wrap items-center gap-x-5 gap-y-2 text-[11px] text-ink-faint lg:justify-end">
                <span className="flex items-center gap-1.5">
                  <Clock3 size={12} />
                  {lastTick.started ? `Last tick ${timeAgo(lastTick.started.created_at)}` : 'No tick recorded'}
                </span>
                {lastTick.durationS !== null && <span>took {lastTick.durationS}s</span>}
                <span>{tickCount} completed ticks</span>
                {repo?.db_size ? (
                  <span className="flex items-center gap-1.5">
                    <Database size={12} /> {formatBytes(repo.db_size)}
                  </span>
                ) : null}
              </div>
            </div>
          </section>

          <section className="grid grid-cols-2 gap-3 lg:grid-cols-3 2xl:grid-cols-6" aria-label="Repository metrics">
            <StatCard
              label="Issues"
              value={summary?.issues.total ?? '—'}
              sub={`${summary?.issues.open ?? 0} open · ${summary?.issues.closed ?? 0} closed`}
              icon={<ListChecks size={15} />}
              accent="#60a5fa"
              loading={summaryLoading}
            />
            <StatCard
              label="In flight"
              value={summary?.issues.in_flight ?? '—'}
              sub="agents currently working"
              icon={<CircleDot size={15} />}
              accent="#a78bfa"
              loading={summaryLoading}
            />
            <StatCard
              label="Active runs"
              value={summary?.runs.running ?? '—'}
              sub={`${summary?.runs.completed ?? 0} completed`}
              icon={<Activity size={15} />}
              accent="#38bdf8"
              loading={summaryLoading}
            />
            <StatCard
              label="Open PRs"
              value={summary?.pulls.open ?? '—'}
              sub={`${summary?.pulls.merged ?? 0} merged all time`}
              icon={<GitPullRequest size={15} />}
              accent="#d878f4"
              loading={summaryLoading}
            />
            <StatCard
              label="Delivered"
              value={summary?.issues.merged ?? '—'}
              sub="issues merged"
              icon={<GitMerge size={15} />}
              accent="#4ed17e"
              loading={summaryLoading}
            />
            <StatCard
              label="Needs attention"
              value={summary?.issues.failed ?? '—'}
              sub="failed or interrupted"
              icon={<TriangleAlert size={15} />}
              accent="#f87171"
              loading={summaryLoading}
              onClick={() => router.push('/issues?stage=failed')}
            />
          </section>

          <div className="grid min-w-0 gap-5 xl:grid-cols-[minmax(0,1.55fr)_minmax(320px,0.72fr)]">
            <div className="min-w-0 space-y-5">
              <Panel
                title="Pipeline"
                action={
                  <span className="text-[10px] font-medium uppercase tracking-[0.1em] text-ink-faint">
                    {summary?.issues.total ?? 0} total issues
                  </span>
                }
              >
                {summaryLoading || !summary ? (
                  <Skeleton className="h-24 w-full" />
                ) : (
                  <PipelineFunnel summary={summary} />
                )}
              </Panel>

              {summary && (
                <FailuresPanel summary={summary} runs={runList} loading={summaryLoading} />
              )}

              <Panel
                title="Recent runs"
                action={
                  <Link href="/runs" className="text-[11px] font-medium text-accent hover:underline">
                    View all <ArrowRight size={11} className="inline" />
                  </Link>
                }
                bodyClassName="p-2 sm:p-3"
              >
                {runs ? <RunsView runs={recentRuns} /> : <Skeleton className="h-48 w-full" />}
              </Panel>
            </div>

            <aside className="min-w-0 space-y-5">
              <Panel
                title="Live activity"
                action={
                  <span className="flex items-center gap-1.5 text-[10px] text-ink-faint">
                    <PulseDot color="#7aaaff" /> polling
                  </span>
                }
                bodyClassName="p-2"
              >
                {events ? (
                  <EventFeed events={events.events.slice(0, 10)} compact />
                ) : (
                  <Skeleton className="h-56 w-full" />
                )}
              </Panel>

              <Panel title="Instance" bodyClassName="p-0">
                <div className="border-b border-edge p-4">
                  <div className="flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] border border-edge bg-panel-2 text-accent">
                      <Bot size={16} />
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-xs font-semibold text-ink">
                        {repo?.agent.model ?? 'Model not reported'}
                      </p>
                      <p className="mt-1 truncate text-[10px] text-ink-faint">
                        {repo?.agent.provider ?? 'Provider not reported'}
                      </p>
                    </div>
                    {repo?.url && (
                      <a
                        href={repo.url}
                        target="_blank"
                        rel="noreferrer"
                        className="btn btn-ghost !h-8 !min-h-8 !w-8 !p-0"
                        aria-label="Open repository"
                        title="Open repository"
                      >
                        <ExternalLink size={14} />
                      </a>
                    )}
                  </div>
                </div>
                <dl className="divide-y divide-edge text-[11px]">
                  <div className="flex items-center justify-between gap-4 px-4 py-3">
                    <dt className="text-ink-faint">Database</dt>
                    <dd className="font-medium text-ink-dim">{repo ? formatBytes(repo.db_size) : '—'}</dd>
                  </div>
                  <div className="flex items-center justify-between gap-4 px-4 py-3">
                    <dt className="text-ink-faint">Bot identity</dt>
                    <dd className="truncate font-medium text-ink-dim">{repo?.agent.bot_user || '—'}</dd>
                  </div>
                  <div className="flex items-center justify-between gap-4 px-4 py-3">
                    <dt className="text-ink-faint">API mode</dt>
                    <dd className="font-medium text-emerald">Read only</dd>
                  </div>
                </dl>
              </Panel>

              <Panel
                title="On the roadmap"
                action={
                  <Link href="/roadmap" className="text-[11px] font-medium text-accent hover:underline">
                    Preview all
                  </Link>
                }
                bodyClassName="p-2"
              >
                <div className="space-y-1">
                  {PLANNED.map(({ icon: Icon, title, description }) => (
                    <div key={title} className="flex gap-3 rounded-xl p-2.5 transition hover:bg-panel-2">
                      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-edge bg-panel-2 text-ink-faint">
                        <Icon size={14} />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="text-[11px] font-semibold text-ink-dim">{title}</p>
                          <span className="rounded-full border border-edge px-1.5 py-0.5 text-[8px] font-semibold uppercase tracking-wide text-ink-faint">
                            Planned
                          </span>
                        </div>
                        <p className="mt-1 text-[10px] leading-4 text-ink-faint">{description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </Panel>
            </aside>
          </div>

          {runList.length === 0 && events?.events.length === 0 && (
            <Panel>
              <EmptyState
                icon={<Sparkles size={19} />}
                title="The control room is ready"
                hint="This instance is healthy but has not recorded an agent run or event yet. New activity will appear here automatically."
                action={
                  <Link href="/issues" className="btn mt-1">
                    Browse issues <ArrowRight size={13} />
                  </Link>
                }
              />
            </Panel>
          )}
        </>
      )}
    </div>
  );
}
