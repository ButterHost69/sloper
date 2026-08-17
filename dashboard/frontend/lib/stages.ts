// Stage metadata for the sloper pipeline.
// The keys mirror the `issues.stage` column values from the Go backend.
export interface StageMeta {
  key: string;
  label: string;
  short: string;
  color: string; // tailwind-friendly hex used inline
  dim?: string;
  desc: string;
}

export const STAGES: StageMeta[] = [
  {
    key: 'new',
    label: 'New',
    short: 'NEW',
    color: '#94a3b8',
    desc: 'Discovered but not yet analyzed',
  },
  {
    key: 'spec-ongoing',
    label: 'Spec Ongoing',
    short: 'SPEC',
    color: '#38bdf8',
    desc: 'Agent is analyzing the issue',
  },
  {
    key: 'spec-done',
    label: 'Spec Done',
    short: 'SPEC',
    color: '#38bdf8',
    desc: 'Spec proposed, awaiting approval',
  },
  {
    key: 'approved',
    label: 'Approved',
    short: 'WORK',
    color: '#a78bfa',
    desc: 'Plan approved, implementing',
  },
  {
    key: 'work-done',
    label: 'Work Done',
    short: 'REVIEW',
    color: '#e879f9',
    desc: 'PR created, self-review in progress',
  },
  {
    key: 'review-done',
    label: 'Review Done',
    short: 'MERGE',
    color: '#fbbf24',
    desc: 'Self-review passed, awaiting merge',
  },
  {
    key: 'merged',
    label: 'Merged',
    short: 'DONE',
    color: '#34d399',
    desc: 'PR merged and cleaned up',
  },
  {
    key: 'failed',
    label: 'Failed',
    short: 'FAIL',
    color: '#f87171',
    desc: 'Stalled or errored',
  },
];

export const STAGE_BY_KEY = Object.fromEntries(STAGES.map((s) => [s.key, s]));

export function stageMeta(key: string): StageMeta {
  return STAGE_BY_KEY[key] ?? {
    key,
    label: key,
    short: key.toUpperCase().slice(0, 5),
    color: '#94a3b8',
    desc: 'Unknown stage',
  };
}

// Pipeline order for the DAG / funnel visualization (failed is a side branch).
export const PIPELINE_ORDER = [
  'new',
  'spec-ongoing',
  'spec-done',
  'approved',
  'work-done',
  'review-done',
  'merged',
] as const;

export interface RunStageMeta {
  key: string;
  label: string;
  color: string;
}

export const RUN_STAGES: Record<string, RunStageMeta> = {
  spec: { key: 'spec', label: 'Spec', color: '#38bdf8' },
  work: { key: 'work', label: 'Work', color: '#a78bfa' },
  review: { key: 'review', label: 'Review', color: '#e879f9' },
  fix: { key: 'fix', label: 'Fix', color: '#fbbf24' },
  merge: { key: 'merge', label: 'Merge', color: '#34d399' },
};

export const RUN_STATUS_COLORS: Record<string, string> = {
  running: '#38bdf8',
  completed: '#34d399',
  failed: '#f87171',
  interrupted: '#fbbf24',
};