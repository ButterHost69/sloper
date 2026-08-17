'use client';

import { useMemo, useState } from 'react';
import Link from 'next/link';
import {
  CheckCircle2,
  CircleDot,
  ExternalLink,
  GitMerge,
  GitPullRequest,
  ListChecks,
  RefreshCw,
  Rocket,
  TriangleAlert,
  XCircle,
} from 'lucide-react';
import { useInstances } from '@/components/instance-context';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { timeAgo, formatBytes } from '@/lib/format';
import { Panel, SectionHeader, StatCard, PulseDot, ErrorState, Skeleton } from '@/components/ui';
import { PipelineFunnel } from '@/components/pipeline-funnel';
import { ActivityChart } from '@/components/activity-chart';
import { EventFeed } from '@/components/event-feed';
import { RunsView } from '@/components/runs-view';

export default function OverviewPage() {
  const { active } = useInstances();
  const [hours, setHours] = useState(24);

  const { data: summary, loading: summaryLoading, error, refresh, refreshing } = useInstanceData(
    (base) => api.summary(base),
    10000,
  );
  const { data: events } = useInstanceData((base) => api.events(base, 40), 10000);
  const { data: runs } = useInstanceData((base) => api.runs(base, 8), 15000);
  const { data: activity } = useInstanceData((base) => api.activity(base, hours), 30000, !!active);
  const { data: health } = useInstanceData((base) => api.health(base), 15000);
  const { data: repo } = useInstanceData((base) => api.repo(base), 60000);

  const eventTypes = useMemo(() => {
    const byType = summary?.events?.by_type ?? {};
    const top = Object.entries(byType)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 6);
    const total = summary?.events?.total ?? 0;
    return { top, total };
  }, [summary]);

  const issueCount = summary?.issues;

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-xs font-medium uppercase tracking-widest text-ink-faint">
            {active ? `Instance · ${active.name}` : 'No active instance'}
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight text-ink">
            {summary?.repo ?? 'Repository'}
          </h1>
          <div className="mt-1.5 flex items-center gap-3 text-xs text-ink-dim">
            {health?.status === 'ok' ? (
              <span className="flex items-center gap-1.5 text-emerald">
                <PulseDot color="#34d399" /> healthy
              </span>
            ) : (
              <span className="flex items-center gap-1.5 text-danger">
                <PulseDot color="#f87171" /> unreachable
              </span>
            )}
            <span>·</span>
            <span>up {health ? timeAgo(new Date(Date.now() - health.uptime_s * 1000).toISOString()) : '—'}</span>
            <span>·</span>
            <span className="flex items-center gap-1">
              <Rocket size={12} /> {summary?.runs?.total ?? 0} runs
            </span>
            {repo?.db_size ? (
              <>
                <span>·</span>
                <span>db {formatBytes(repo.db_size)}</span>
              </>
            ) : null}
            {repo?.url ? (
              <>
                <span>·</span>
                <a
                  href={repo.url}
                  target="_blank"
                  rel="noreferrer"
                  className="flex items-center gap-1 text-accent hover:underline"
                >
                  repo <ExternalLink size={11} />
                </a>
              </>
            ) : null}
          </div>
        </div>
        <button className="btn" onClick={refresh} disabled={refreshing || summaryLoading}>
          <RefreshCw size={14} className={refreshing ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      {error ? (
        <Panel>
          <ErrorState error={error} onRetry={refresh} />
        </Panel>
      ) : (
        <>
          {/* Stat cards */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-6">
            <StatCard
              label="Issues"
              value={issueCount?.total ?? '—'}
              sub={`${issueCount?.open ?? 0} open · ${issueCount?.closed ?? 0} closed`}
              icon={<ListChecks size={16} />}
              accent="#94a3b8"
              loading={summaryLoading}
            />
            <StatCard
              label="In flight"
              value={issueCount?.in_flight ?? '—'}
              sub="agents working"
              icon={<CircleDot size={16} />}
              accent="#a78bfa"
              loading={summaryLoading}
            />
            <StatCard
              label="Merged"
              value={issueCount?.merged ?? '—'}
              sub="delivered"
              icon={<GitMerge size={16} />}
              accent="#34d399"
              loading={summaryLoading}
            />
            <StatCard
              label="Failed"
              value={issueCount?.failed ?? '—'}
              sub="need attention"
              icon={<TriangleAlert size={16} />}
              accent="#f87171"
              loading={summaryLoading}
              onClick={() => (window.location.href = '/issues?stage=failed')}
            />
            <StatCard
              label="Active runs"
              value={summary?.runs?.running ?? '—'}
              sub={`${summary?.runs?.completed ?? 0} completed`}
              icon={<RefreshCw size={16} />}
              accent="#38bdf8"
              loading={summaryLoading}
            />
            <StatCard
              label="Open PRs"
              value={summary?.pulls?.open ?? '—'}
              sub={`${summary?.pulls?.total ?? 0} total`}
              icon={<GitPullRequest size={16} />}
              accent="#e879f9"
              loading={summaryLoading}
            />
          </div>

          {/* Pipeline funnel */}
          <Panel title="Pipeline">
            {summaryLoading ? (
              <div className="space-y-3">
                <Skeleton className="h-24 w-full" />
              </div>
            ) : (
              <PipelineFunnel summary={summary!} />
            )}
          </Panel>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-5">
            {/* Activity chart */}
            <div className="xl:col-span-3">
              <Panel
                title="Activity"
                action={
                  <div className="flex gap-1">
                    {[6, 24, 168].map((h) => (
                      <button
                        key={h}
                        onClick={() => setHours(h)}
                        className={`rounded-md px-2 py-1 text-[0.65rem] font-semibold transition ${
                          hours === h
                            ? 'bg-accent/15 text-accent'
                            : 'text-ink-faint hover:text-ink'
                        }`}
                      >
                        {h === 168 ? '7d' : `${h}h`}
                      </button>
                    ))}
                  </div>
                }
              >
                {activity ? (
                  <ActivityChart buckets={activity.buckets} />
                ) : (
                  <Skeleton className="h-[220px] w-full" />
                )}
              </Panel>

              {/* Runs */}
              <div className="mt-4">
                <SectionHeader
                  title="Recent runs"
                  action={
                    <Link href="/runs" className="text-xs text-accent hover:underline">
                      View all →
                    </Link>
                  }
                />
                {runs ? (
                  <RunsView runs={runs.runs} defaultOpenFirst />
                ) : (
                  <Skeleton className="h-48 w-full" />
                )}
              </div>
            </div>

            {/* Events */}
            <div className="xl:col-span-2">
              <SectionHeader
                title="Live events"
                action={
                  <Link href="/events" className="text-xs text-accent hover:underline">
                    View all →
                  </Link>
                }
              />
              <Panel bodyClassName="p-2">
                <EventFeed events={events?.events ?? []} compact />
              </Panel>

              {/* Event type distribution */}
              <div className="mt-4">
                <SectionHeader title="Event mix" />
                <Panel>
                  {eventTypes.total === 0 ? (
                    <p className="py-4 text-center text-xs text-ink-faint">No events recorded</p>
                  ) : (
                    <div className="space-y-2.5">
                      {eventTypes.top.map(([type, count]) => {
                        const pct = Math.round((count / eventTypes.total) * 100);
                        return (
                          <div key={type}>
                            <div className="mb-1 flex items-center justify-between text-xs">
                              <span className="font-mono text-ink-dim">{type}</span>
                              <span className="text-ink-faint">
                                {count} · {pct}%
                              </span>
                            </div>
                            <div className="h-1.5 overflow-hidden rounded-full bg-edge">
                              <div
                                className="h-full rounded-full"
                                style={{
                                  width: `${pct}%`,
                                  background:
                                    type.includes('failed') || type.includes('error')
                                      ? '#f87171'
                                      : '#34d399',
                                }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </Panel>
              </div>
            </div>
          </div>

          {/* Bottom strip */}
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <StatCard
              label="Spec runs"
              value={summary?.runs?.by_status?.completed ?? '—'}
              sub="completed"
              icon={<CheckCircle2 size={16} />}
              accent="#38bdf8"
              loading={summaryLoading}
            />
            <StatCard
              label="Interrupted"
              value={summary?.runs?.interrupted ?? '—'}
              sub="recovered on boot"
              icon={<XCircle size={16} />}
              accent="#fbbf24"
              loading={summaryLoading}
            />
            <StatCard
              label="Events"
              value={summary?.events?.total ?? '—'}
              sub="audit log entries"
              icon={<CircleDot size={16} />}
              accent="#94a3b8"
              loading={summaryLoading}
            />
            <StatCard
              label="Failed runs"
              value={summary?.runs?.failed ?? '—'}
              sub="errored executions"
              icon={<TriangleAlert size={16} />}
              accent="#f87171"
              loading={summaryLoading}
              onClick={() => (window.location.href = '/runs?status=failed')}
            />
          </div>
        </>
      )}
    </div>
  );
}