import type { Session, SessionEntry, SessionGroup, SessionPR, SessionToolRef } from './types';

/** A one-line answer to "what is this worker doing right now". */
export interface SessionActivity {
  label: string;
  detail?: string;
  tone: 'live' | 'idle' | 'error' | 'waiting';
}

export function describeCurrent(session: Session): SessionActivity {
  const current = session.summary.current;
  const live = session.live;
  if (!current) {
    return { label: live ? 'Starting up' : 'No activity yet', tone: 'idle' };
  }
  switch (current.kind) {
    case 'tool': {
      const tools = current.tools ?? [];
      const first = tools[0];
      const label = tools.length > 1
        ? `${tools.length} tools ${live ? 'running' : 'left unfinished'}`
        : `${first?.name ?? 'tool'} ${live ? 'running' : 'left unfinished'}`;
      return {
        label,
        detail: first?.preview || undefined,
        tone: live ? 'live' : 'idle',
      };
    }
    case 'thinking':
      // "Model turn finished" is only interesting while the worker is still
      // writing; a session that went quiet is simply idle.
      return live
        ? { label: current.text ?? 'Thinking', tone: 'live' }
        : { label: 'Idle', tone: 'idle' };
    case 'waiting':
      return { label: current.text ?? 'Waiting for the model', tone: 'waiting' };
    case 'starting':
      return { label: current.text ?? 'Starting', tone: 'idle' };
    case 'error':
      return { label: current.text ?? 'Model error', tone: 'error' };
    default:
      return { label: current.text ?? current.kind, tone: 'idle' };
  }
}

export function toolLabel(tool: SessionToolRef): string {
  return tool.preview ? `${tool.name} · ${tool.preview}` : tool.name;
}

export function formatTokens(tokens: number | undefined): string {
  if (!tokens) return '—';
  if (tokens < 1000) return String(tokens);
  if (tokens < 1_000_000) return `${(tokens / 1000).toFixed(tokens < 10_000 ? 1 : 0)}k`;
  return `${(tokens / 1_000_000).toFixed(1)}M`;
}

export function formatCost(cost: number | undefined): string {
  if (!cost) return '—';
  if (cost < 0.01) return `$${cost.toFixed(4)}`;
  return `$${cost.toFixed(2)}`;
}

/**
 * How long the worker has been on its current step. Only meaningful for a
 * pending tool call, where the file gives us the assistant message that
 * requested it.
 */
export function currentStepMs(session: Session): number | null {
  const since = session.summary.current?.since;
  if (!since) return null;
  const started = new Date(since).getTime();
  if (Number.isNaN(started)) return null;
  const ms = Date.now() - started;
  return ms > 0 ? ms : null;
}

export function formatElapsed(ms: number): string {
  const s = Math.floor(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ${s % 60}s`;
  const h = Math.floor(m / 60);
  return `${h}h ${m % 60}m`;
}

export function entryKey(entry: SessionEntry, index: number): string {
  return entry.id ? `${entry.type}:${entry.id}` : `${entry.type}:${entry.timestamp ?? ''}:${index}`;
}

export function isMessage(entry: SessionEntry): boolean {
  return entry.type === 'message' && !!entry.message;
}

/**
 * A pull-request view of the same sessions: every group is keyed by the PR an
 * issue produced, with issues that have no PR collected in one bucket.
 */
export interface PRGroup {
  key: string;
  pr: SessionPR | null;
  issueNumbers: number[];
  title: string;
  live_count: number;
  updated_at: string;
  sessions: Session[];
}

export function groupSessionsByPR(groups: SessionGroup[]): PRGroup[] {
  const byPR = new Map<string, PRGroup>();

  for (const group of groups) {
    const pr = group.pr ?? null;
    const unscoped = group.issue_number === 0;
    // Sessions with no issue at all are not "waiting for a PR" — keep them apart.
    const key = pr ? `pr-${pr.number}` : unscoped ? 'unscoped' : 'no-pr';
    let target = byPR.get(key);
    if (!target) {
      target = {
        key,
        pr,
        issueNumbers: [],
        title: pr
          ? group.title ?? ''
          : unscoped
            ? 'Sessions outside the pipeline'
            : 'Issues without a pull request',
        live_count: 0,
        updated_at: '',
        sessions: [],
      };
      byPR.set(key, target);
    }
    if (group.issue_number > 0 && !target.issueNumbers.includes(group.issue_number)) {
      target.issueNumbers.push(group.issue_number);
    }
    if (!target.title && group.title) target.title = group.title;
    target.live_count += group.live_count;
    if ((group.updated_at ?? '') > target.updated_at) target.updated_at = group.updated_at ?? '';
    target.sessions.push(...group.sessions);
  }

  const out = [...byPR.values()];
  for (const group of out) {
    group.sessions.sort((a, b) => (a.started_at ?? a.modified_at).localeCompare(b.started_at ?? b.modified_at));
  }
  out.sort((a, b) => {
    if (a.live_count !== b.live_count) return b.live_count - a.live_count;
    return (b.updated_at ?? '').localeCompare(a.updated_at ?? '');
  });
  return out;
}
