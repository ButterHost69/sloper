import type { AttachedRepo, ControllerHealth } from './types';
import { normalizeUrl } from './instances';

const STORAGE_KEY = 'sloper.controller.v1';

/** Controller URL the Repositories page starts with. */
export const DEFAULT_CONTROLLER_URL =
  process.env.NEXT_PUBLIC_SLOPER_CONTROLLER_URL?.trim() || 'http://localhost:9090';

export interface ControllerConfig {
  url: string;
  token?: string;
}

export class ControllerError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export function loadController(): ControllerConfig {
  const fallback: ControllerConfig = { url: normalizeUrl(DEFAULT_CONTROLLER_URL) };
  if (typeof window === 'undefined') return fallback;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return fallback;
    const parsed = JSON.parse(raw) as ControllerConfig;
    if (!parsed?.url) return fallback;
    return { url: normalizeUrl(parsed.url), token: parsed.token || undefined };
  } catch {
    return fallback;
  }
}

export function saveController(config: ControllerConfig): void {
  if (typeof window === 'undefined') return;
  const next: ControllerConfig = {
    url: normalizeUrl(config.url),
    token: config.token?.trim() || undefined,
  };
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
}

/**
 * Talks to the sloper-controller. The controller is the only backend the
 * console writes to: attaching a repo there starts a container for it.
 */
async function request<T>(
  config: ControllerConfig,
  path: string,
  init: RequestInit = {},
  timeoutMs = 20000,
): Promise<T> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  const token = config.token?.trim();

  try {
    const res = await fetch(`${normalizeUrl(config.url)}${path}`, {
      ...init,
      signal: controller.signal,
      headers: {
        Accept: 'application/json',
        ...(init.body ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(init.headers ?? {}),
      },
    });
    if (!res.ok) {
      let message = `HTTP ${res.status}`;
      try {
        const body = (await res.json()) as { error?: string };
        if (body.error) message = body.error;
      } catch {
        /* keep the status line */
      }
      throw new ControllerError(res.status, message);
    }
    return (await res.json()) as T;
  } catch (err) {
    if (err instanceof ControllerError) throw err;
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new ControllerError(0, 'The controller did not answer in time');
    }
    throw new ControllerError(
      0,
      err instanceof Error ? err.message : 'Controller unreachable',
    );
  } finally {
    clearTimeout(timer);
  }
}

export const controllerApi = {
  health: (cfg: ControllerConfig) => request<ControllerHealth>(cfg, '/api/health', {}, 6000),

  listRepos: (cfg: ControllerConfig) =>
    request<{ repos: AttachedRepo[] | null; count: number }>(cfg, '/api/repos'),

  /** Attach a repo by link; the controller starts its container. */
  attachRepo: (cfg: ControllerConfig, link: string) =>
    request<AttachedRepo>(
      cfg,
      '/api/repos',
      { method: 'POST', body: JSON.stringify({ link }) },
      30000,
    ),

  setDesiredState: (cfg: ControllerConfig, id: number, desired: 'running' | 'stopped') =>
    request<AttachedRepo>(
      cfg,
      `/api/repos/${id}`,
      { method: 'PATCH', body: JSON.stringify({ desired_state: desired }) },
      30000,
    ),

  restartRepo: (cfg: ControllerConfig, id: number) =>
    request<AttachedRepo>(cfg, `/api/repos/${id}/restart`, { method: 'POST' }, 60000),

  /** Detach removes the container; purge also deletes the repo's volumes. */
  detachRepo: (cfg: ControllerConfig, id: number, purge = false) =>
    request<{ status: string; id: number; purged: boolean }>(
      cfg,
      `/api/repos/${id}${purge ? '?purge=true' : ''}`,
      { method: 'DELETE' },
      60000,
    ),

  logs: (cfg: ControllerConfig, id: number, tail = 200) =>
    request<{ id: number; tail: number; logs: string }>(
      cfg,
      `/api/repos/${id}/logs?tail=${tail}`,
    ),
};
