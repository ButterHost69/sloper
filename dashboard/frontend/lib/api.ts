import type {
  ActivityBucket,
  CommentRecord,
  EventRecord,
  Health,
  Issue,
  IssueDetail,
  PullRecord,
  RepoInfo,
  RunRecord,
  Summary,
} from './types';
import { normalizeUrl, tokenFor } from './instances';

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(base: string, path: string, timeoutMs = 8000): Promise<T> {
  const url = `${normalizeUrl(base)}${path}`;
  const token = tokenFor(base);
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(url, {
      signal: controller.signal,
      headers: {
        Accept: 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });
    if (!res.ok) {
      let msg = `HTTP ${res.status}`;
      try {
        const body = (await res.json()) as { error?: string };
        if (body.error) msg = body.error;
      } catch {
        /* ignore */
      }
      throw new ApiError(res.status, msg);
    }
    return (await res.json()) as T;
  } catch (err) {
    if (err instanceof ApiError) throw err;
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new ApiError(0, 'Request timed out');
    }
    throw new ApiError(0, err instanceof Error ? err.message : 'Network error');
  } finally {
    clearTimeout(timer);
  }
}

export interface IssuesParams {
  limit?: number;
  stage?: string;
  q?: string;
}

export const api = {
  health: (base: string) => request<Health>(base, '/api/health', 5000),
  repo: (base: string) => request<RepoInfo>(base, '/api/repo'),
  summary: (base: string) => request<Summary>(base, '/api/summary'),
  issues: (base: string, params: IssuesParams = {}) => {
    const sp = new URLSearchParams();
    if (params.limit) sp.set('limit', String(params.limit));
    if (params.stage && params.stage !== 'all') sp.set('stage', params.stage);
    if (params.q) sp.set('q', params.q);
    const qs = sp.toString();
    return request<{ issues: Issue[]; count: number }>(base, `/api/issues${qs ? `?${qs}` : ''}`);
  },
  issueDetail: (base: string, number: number) =>
    request<IssueDetail>(base, `/api/issues/${number}`),
  comments: (base: string, number: number) =>
    request<{ comments: CommentRecord[] }>(base, `/api/issues/${number}/comments`),
  runs: (base: string, limit = 500) =>
    request<{ runs: RunRecord[]; count: number }>(base, `/api/runs?limit=${limit}`),
  pulls: (base: string, limit = 500) =>
    request<{ pulls: PullRecord[]; count: number }>(base, `/api/pulls?limit=${limit}`),
  events: (base: string, limit = 300) =>
    request<{ events: EventRecord[]; count: number }>(base, `/api/events?limit=${limit}`),
  activity: (base: string, hours = 24) =>
    request<{ hours: number; buckets: ActivityBucket[] }>(
      base,
      `/api/metrics/activity?hours=${hours}`,
    ),
};