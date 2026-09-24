import Link from 'next/link';
import {
  Activity,
  ArrowRight,
  Box,
  Check,
  CircleDashed,
  Database,
  GitPullRequest,
  Layers3,
  ListChecks,
  LockKeyhole,
  Network,
  ScrollText,
  ServerCog,
  Settings2,
  Sparkles,
  Store,
} from 'lucide-react';
import { Badge, Panel } from '@/components/ui';

const liveFeatures = [
  {
    title: 'Issues',
    description: 'Search cached issues, follow pipeline stage, and inspect agent specs and activity.',
    href: '/issues',
    icon: ListChecks,
  },
  {
    title: 'Pull Requests',
    description: 'Track open PRs, review rounds, and the point where a change is ready for human acceptance.',
    href: '/pulls',
    icon: GitPullRequest,
  },
  {
    title: 'Runs',
    description: 'Monitor in-flight and failed agent runs with stage, status, and execution history.',
    href: '/runs',
    icon: Activity,
  },
  {
    title: 'Events',
    description: 'Inspect the read-only audit stream and hourly activity timeline for the selected range.',
    href: '/events',
    icon: ScrollText,
  },
  {
    title: 'Instances',
    description: 'Connect to multiple read-only sloper APIs and inspect their health and repository context.',
    href: '/instances',
    icon: ServerCog,
  },
];

const plannedFeatures = [
  {
    title: 'Condition-based loops',
    eyebrow: 'Automation',
    description:
      'Define repeatable loops that branch on task, review, or test conditions instead of following one fixed chain.',
    detail: 'Custom, semi-automated, and fully automated loop recipes are unavailable.',
    icon: Settings2,
    color: '#a78bfa',
    href: '/runs',
    linkLabel: 'See today’s fixed pipeline',
  },
  {
    title: 'Remote runner pools',
    eyebrow: 'Execution',
    description:
      'Route work to remote hosts and containerized environments when an agent needs a different test surface.',
    detail: 'Runner discovery, allocation, and container lifecycle controls are unavailable.',
    icon: Box,
    color: '#38bdf8',
    href: '/instances',
    linkLabel: 'View registered instances',
  },
  {
    title: 'Extension & loop marketplace',
    eyebrow: 'Ecosystem',
    description:
      'Share reusable agents, extensions, and workflows so teams can discover and install proven loop packages.',
    detail: 'Publishing, discovery, versioning, and installation are unavailable.',
    icon: Store,
    color: '#e879f9',
    href: '/pulls',
    linkLabel: 'Browse current pull requests',
  },
  {
    title: 'Project & branch memory',
    eyebrow: 'Context',
    description:
      'Preserve project-wide decisions and branch-specific context across long review and fix cycles.',
    detail: 'Persistent memory scopes and retrieval across sessions are unavailable.',
    icon: Database,
    color: '#34d399',
    href: '/issues',
    linkLabel: 'Explore issue context',
  },
];

const pipelineStages = [
  { label: 'Issue', icon: ListChecks, state: 'live' },
  { label: 'Spec', icon: Sparkles, state: 'live' },
  { label: 'Work', icon: Layers3, state: 'live' },
  { label: 'Review', icon: Check, state: 'live' },
  { label: 'Loop', icon: Settings2, state: 'planned' },
  { label: 'Remote run', icon: Network, state: 'planned' },
] as const;

export default function RoadmapPage() {
  return (
    <div className="space-y-8">
      <header className="glass-hero panel relative overflow-hidden px-5 py-6 sm:px-7 sm:py-7">
        <div className="relative max-w-3xl">
          <div className="flex flex-wrap items-center gap-2">
            <Badge color="#34d399" dot>
              Live observability
            </Badge>
            <Badge color="#a78bfa">Planned capabilities</Badge>
          </div>
          <h1 className="mt-4 text-2xl font-bold tracking-tight text-ink sm:text-3xl">
            From a working agent pipeline to a programmable agent platform
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-ink-dim">
            Sloper’s GitHub-centered workflow is live today. The items below are a product
            roadmap—not available console controls—and show where custom automation, remote
            execution, reusable workflows, and durable context are headed.
          </p>
          <div className="mt-5 flex flex-wrap gap-x-5 gap-y-2 text-xs text-ink-faint">
            <span className="flex items-center gap-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-accent" /> Available now
            </span>
            <span className="flex items-center gap-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-violet" /> Planned
            </span>
            <span className="flex items-center gap-1.5">
              <LockKeyhole size={12} /> Read-only dashboard preview
            </span>
          </div>
        </div>
      </header>

      <section aria-labelledby="live-heading" className="space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-2">
          <div>
            <p className="text-xs font-semibold uppercase tracking-widest text-accent">Shipped</p>
            <h2 id="live-heading" className="mt-1 text-xl font-bold tracking-tight text-ink">
              Available capabilities
            </h2>
          </div>
          <p className="max-w-md text-xs leading-5 text-ink-faint">
            These routes are connected to the running console and reflect data from your registered sloper instances.
          </p>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {liveFeatures.map(({ title, description, href, icon: Icon }) => (
            <Link
              key={title}
              href={href}
              className="panel panel-hover group flex min-h-44 flex-col p-4 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
            >
              <div className="flex items-start justify-between gap-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-lg border border-accent/25 bg-accent/10 text-accent">
                  <Icon size={17} aria-hidden="true" />
                </span>
                <Badge color="var(--color-accent)" dot>
                  Live
                </Badge>
              </div>
              <h3 className="mt-4 text-sm font-semibold text-ink group-hover:text-accent">{title}</h3>
              <p className="mt-1.5 flex-1 text-xs leading-5 text-ink-faint">{description}</p>
              <span className="mt-4 flex items-center gap-1 text-xs font-medium text-accent">
                Open {title.toLowerCase()} <ArrowRight size={12} aria-hidden="true" />
              </span>
            </Link>
          ))}
        </div>
      </section>

      <section aria-labelledby="planned-heading" className="space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-2">
          <div>
            <p className="text-xs font-semibold uppercase tracking-widest text-violet">Future direction</p>
            <h2 id="planned-heading" className="mt-1 text-xl font-bold tracking-tight text-ink">
              Planned capabilities
            </h2>
          </div>
          <p className="max-w-md text-xs leading-5 text-ink-faint">
            No configuration or installation flow exists for these items yet.
          </p>
        </div>

        <div className="grid gap-4 lg:grid-cols-2">
          {plannedFeatures.map(
            ({ title, eyebrow, description, detail, icon: Icon, color, href, linkLabel }) => (
              <article key={title} className="panel relative overflow-hidden p-5 sm:p-6">
              <div className="relative">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <span
                      className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border"
                      style={{ color, borderColor: `${color}35`, backgroundColor: `${color}10` }}
                    >
                      <Icon size={19} aria-hidden="true" />
                    </span>
                    <div>
                      <p className="text-[0.65rem] font-semibold uppercase tracking-widest text-ink-faint">{eyebrow}</p>
                      <h3 className="mt-0.5 text-base font-semibold text-ink">{title}</h3>
                    </div>
                  </div>
                  <Badge color={color}>
                    Planned
                  </Badge>
                </div>
                <p className="mt-4 text-sm leading-6 text-ink-dim">{description}</p>
                <div className="mt-4 rounded-lg border border-violet/20 bg-violet/[0.05] p-3">
                  <p className="flex items-start gap-2 text-xs leading-5 text-ink-dim">
                    <CircleDashed size={14} className="mt-0.5 shrink-0 text-violet" aria-hidden="true" />
                    <span>
                      <strong className="font-semibold text-ink">Not implemented yet.</strong>{' '}
                      {detail}
                    </span>
                  </p>
                </div>
                <Link
                  href={href}
                  className="hit-target mt-4 inline-flex items-center gap-1.5 text-xs font-medium text-accent hover:underline"
                >
                  {linkLabel} <ArrowRight size={12} aria-hidden="true" />
                </Link>
              </div>
              </article>
            ),
          )}
        </div>
      </section>

      <section aria-labelledby="pipeline-heading" className="space-y-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-ink-faint">Concept only</p>
          <h2 id="pipeline-heading" className="mt-1 text-xl font-bold tracking-tight text-ink">
            Future orchestration pipeline
          </h2>
          <p className="mt-1 max-w-2xl text-xs leading-5 text-ink-faint">
            This disabled preview is illustrative only. The live console observes fixed sloper runs; it does not configure this graph.
          </p>
        </div>

        <Panel className="overflow-hidden">
          <div className="flex flex-wrap items-center justify-between gap-3 border-b border-edge bg-panel-2/60 px-4 py-3">
            <div>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-ink-dim">Condition-aware execution</h3>
              <p className="mt-0.5 text-[0.68rem] text-ink-faint">Planned · architecture preview, not a working control</p>
            </div>
            <Badge color="#a78bfa">Planned</Badge>
          </div>

          <div aria-label="Planned orchestration pipeline preview" className="px-4 py-5 sm:px-6 sm:py-6">
            <ol className="grid gap-3 md:grid-cols-3 xl:grid-cols-6">
              {pipelineStages.map(({ label, icon: Icon, state }, index) => (
                <li key={label} className="relative">
                  <div
                    className={`flex h-full min-h-28 flex-col items-center justify-center rounded-xl border px-3 py-4 text-center ${
                      state === 'live'
                        ? 'border-accent/20 bg-accent/[0.04]'
                        : 'border-dashed border-violet/35 bg-violet/[0.04]'
                    }`}
                  >
                    <Icon
                      size={18}
                      aria-hidden="true"
                      className={state === 'live' ? 'text-accent' : 'text-violet'}
                    />
                    <span className="mt-2 text-xs font-semibold text-ink">{label}</span>
                    <span
                      className={`mt-1 text-[0.6rem] font-semibold uppercase tracking-wider ${
                        state === 'live' ? 'text-accent' : 'text-violet'
                      }`}
                    >
                      {state === 'live' ? 'Live stage' : 'Planned'}
                    </span>
                  </div>
                  {index < pipelineStages.length - 1 ? (
                    <ArrowRight
                      size={14}
                      className="absolute -right-[0.65rem] top-1/2 z-10 hidden -translate-y-1/2 text-ink-faint xl:block"
                      aria-hidden="true"
                    />
                  ) : null}
                </li>
              ))}
            </ol>
            <div className="mt-5 flex flex-wrap items-center justify-between gap-3 border-t border-edge pt-4">
              <p className="text-xs leading-5 text-ink-faint">
                Future loops could branch after review, request a remote runner, and consult shared project memory before continuing.
              </p>
              <span className="chip border-violet/30 bg-violet/10 text-violet">
                <LockKeyhole size={11} aria-hidden="true" /> Preview disabled
              </span>
            </div>
          </div>
        </Panel>
      </section>
    </div>
  );
}
