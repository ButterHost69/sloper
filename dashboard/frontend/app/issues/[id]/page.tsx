'use client';

import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  ExternalLink,
  FileDiff,
  GitPullRequest,
  ListChecks,
  MessageSquare,
  User,
} from 'lucide-react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import { formatTime, timeAgo, truncate } from '@/lib/format';
import { ErrorState, Panel, Skeleton } from '@/components/ui';
import { LabelChips, ReviewStateBadge, StageBadge } from '@/components/badges';
import { FailedMarker, PipelineDAG } from '@/components/pipeline-dag';
import { RunsView } from '@/components/runs-view';
import { EventFeed } from '@/components/event-feed';
import type { Spec } from '@/lib/types';

function SpecPanel({ spec }: { spec: Spec | null }) {
  if (!spec) {
    return (
      <p className="py-6 text-center text-xs text-ink-faint">
        No cached spec for this issue yet.
      </p>
    );
  }
  return (
    <div className="space-y-4">
      <div>
        <h4 className="mb-1.5 flex items-center gap-1.5 text-[0.7rem] font-semibold uppercase tracking-wider text-ink-faint">
          <ListChecks size={13} /> Summary
        </h4>
        <p className="text-sm leading-relaxed text-ink-dim">{spec.summary || '—'}</p>
      </div>
      <div>
        <h4 className="mb-1.5 flex items-center gap-1.5 text-[0.7rem] font-semibold uppercase tracking-wider text-ink-faint">
          <FileDiff size={13} /> Files to change
        </h4>
        {spec.files_to_change?.length ? (
          <div className="flex flex-wrap gap-1.5">
            {spec.files_to_change.map((f) => (
              <code key={f} className="rounded border border-edge-2 bg-panel-2 px-2 py-0.5 font-mono text-[0.7rem] text-info">
                {f}
              </code>
            ))}
          </div>
        ) : (
          <p className="text-xs text-ink-faint">Not determined yet.</p>
        )}
      </div>
      <div>
        <h4 className="mb-1.5 text-[0.7rem] font-semibold uppercase tracking-wider text-ink-faint">
          Implementation plan
        </h4>
        <div className="prose-dark max-h-72 overflow-auto rounded-lg border border-edge bg-[#0c0c0f] p-3">
          {spec.implementation_plan
            .split('\n')
            .map((line, i) => <p key={i} className="mb-1.5">{line || '\u00A0'}</p>)}
        </div>
      </div>
    </div>
  );
}

export default function IssueDetailPage() {
  const params = useParams<{ id: string }>();
  const number = Number(params.id);
  const { data, error, refresh } = useInstanceData(
    (base) => api.issueDetail(base, number),
    10000,
  );

  if (!data) {
    return (
      <div className="space-y-5">
        <Link href="/issues" className="flex items-center gap-1.5 text-xs text-ink-dim hover:text-ink">
          <ArrowLeft size={14} /> Back to issues
        </Link>
        {error ? (
          <Panel>
            <ErrorState error={error} onRetry={refresh} />
          </Panel>
        ) : (
          <>
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-64 w-full" />
          </>
        )}
      </div>
    );
  }

  const issue = data!.issue;
  const meta = data!.spec;

  return (
    <div className="space-y-5">
      <Link href="/issues" className="flex w-fit items-center gap-1.5 text-xs text-ink-dim transition hover:text-ink">
        <ArrowLeft size={14} /> Back to issues
      </Link>

      {/* Header */}
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2.5">
            <span className="font-mono text-sm text-ink-faint">#{issue.number}</span>
            <StageBadge stage={issue.stage} />
            {issue.pr_number > 0 && (
              <ReviewStateBadge state={data!.pr?.review_state ?? 'none'} />
            )}
          </div>
          <h1 className="mt-2 text-2xl font-bold tracking-tight text-ink">{issue.title}</h1>
          <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-ink-dim">
            <span className="flex items-center gap-1">
              <User size={12} /> @{issue.author}
            </span>
            <span>created {timeAgo(issue.created_at)}</span>
            <span>updated {timeAgo(issue.updated_at_local)}</span>
            {issue.review_iterations > 0 && <span>{issue.review_iterations} review round(s)</span>}
            <LabelChips labels={issue.labels} />
          </div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          {issue.url && (
            <a href={issue.url} target="_blank" rel="noreferrer" className="btn">
              Issue <ExternalLink size={13} />
            </a>
          )}
          {data!.pr?.url && (
            <a href={data!.pr.url} target="_blank" rel="noreferrer" className="btn">
              PR #{data!.pr.number} <ExternalLink size={13} />
            </a>
          )}
        </div>
      </div>

      {/* DAG */}
      <Panel title="Pipeline">
        <PipelineDAG stage={issue.stage} />
        <FailedMarker stage={issue.stage} />
      </Panel>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        {/* Spec */}
        <Panel title="Spec analysis" bodyClassName="p-5">
          <SpecPanel spec={meta} />
        </Panel>

        {/* Runs timeline */}
        <Panel
          title="Runs"
          action={
            <span className="text-xs text-ink-faint">{data!.runs.length} executions</span>
          }
          bodyClassName="p-2"
        >
          <RunsView runs={data!.runs} defaultOpenFirst />
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        {/* Comments */}
        <Panel title="Comments" bodyClassName="p-4">
          {data!.comments.length === 0 ? (
            <p className="py-6 text-center text-xs text-ink-faint">No cached comments.</p>
          ) : (
            <div className="space-y-3">
              {data!.comments.map((c) => (
                <div key={c.id} className="rounded-lg border border-edge bg-panel-2 p-3">
                  <div className="mb-1.5 flex items-center gap-2 text-xs">
                    <span className="flex items-center gap-1 font-semibold text-ink">
                      <User size={11} /> {c.author}
                    </span>
                    <span className="text-ink-faint">{formatTime(c.created_at)}</span>
                    {c.processed && (
                      <span className="chip !text-[0.58rem] !text-emerald !border-emerald/30">
                        processed
                      </span>
                    )}
                    {c.replied_by_bot && (
                      <span className="chip !text-[0.58rem] !text-fuchsia !border-fuchsia/30">
                        bot replied
                      </span>
                    )}
                  </div>
                  <p className="whitespace-pre-wrap text-sm leading-relaxed text-ink-dim">
                    {truncate(c.body, 600)}
                  </p>
                </div>
              ))}
            </div>
          )}
        </Panel>

        {/* Events */}
        <Panel title="Event timeline" bodyClassName="p-2">
          <EventFeed events={data!.events} />
        </Panel>
      </div>

      {/* PR info */}
      {data!.pr && (
        <Panel title="Linked pull request" bodyClassName="p-4">
          <div className="flex flex-wrap items-center gap-4">
            <span className="flex items-center gap-2 text-sm font-medium text-ink">
              <GitPullRequest size={15} className="text-fuchsia" />
              <a
                href={data!.pr.url}
                target="_blank"
                rel="noreferrer"
                className="hover:text-accent hover:underline"
              >
                {data!.pr.title}
              </a>
            </span>
            <ReviewStateBadge state={data!.pr.review_state} />
            <code className="font-mono text-xs text-ink-faint">
              {data!.pr.head_sha.slice(0, 8)}
              <span className="text-ink-faint/50"> → </span>
              {data!.pr.base_sha.slice(0, 8)}
            </code>
            <span className="ml-auto flex items-center gap-1.5 text-xs text-ink-faint">
              <MessageSquare size={12} /> created {timeAgo(data!.pr.updated_at)}
            </span>
          </div>
        </Panel>
      )}
    </div>
  );
}