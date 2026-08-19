import type { Instance } from './types';

const STORAGE_KEY = 'sloper.instances.v1';
export const ACTIVE_KEY = 'sloper.activeInstance.v1';

const DEFAULT_INSTANCE: Instance = {
  id: 'local',
  name: 'Local Sloper',
  url: 'http://localhost:8080',
};

/**
 * Registry of instance base URL -> bearer token, kept in sync with the
 * persisted instance list so the API client can attach credentials without
 * threading a token through every call site.
 */
const tokenRegistry = new Map<string, string>();

function registerTokens(instances: Instance[]): void {
  tokenRegistry.clear();
  for (const inst of instances) {
    tokenRegistry.set(normalizeUrl(inst.url), inst.token ?? '');
  }
}

export function tokenFor(url: string): string {
  return tokenRegistry.get(normalizeUrl(url)) ?? '';
}

export function loadInstances(): Instance[] {
  if (typeof window === 'undefined') return [DEFAULT_INSTANCE];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [DEFAULT_INSTANCE];
    const parsed = JSON.parse(raw) as Instance[];
    if (!Array.isArray(parsed) || parsed.length === 0) return [DEFAULT_INSTANCE];
    registerTokens(parsed);
    return parsed;
  } catch {
    return [DEFAULT_INSTANCE];
  }
}

export function saveInstances(instances: Instance[]): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(instances));
  registerTokens(instances);
}

export function loadActiveInstanceId(): string | null {
  if (typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(ACTIVE_KEY);
  } catch {
    return null;
  }
}

export function saveActiveInstanceId(id: string): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(ACTIVE_KEY, id);
}

/**
 * Normalizes an instance base URL: trims surrounding whitespace, strips
 * trailing slashes, and prepends `http://` when no scheme is present so that
 * `new URL()` and `fetch()` never fail on user-entered values.
 */
export function normalizeUrl(url: string): string {
  let u = url.trim().replace(/\/+$/, '');
  if (!/^[a-z][a-z0-9+.-]*:\/\//i.test(u)) {
    u = `http://${u}`;
  }
  return u;
}

/** Returns the `host:port` of an instance URL, or '' if it cannot be parsed. */
export function safeHost(url: string): string {
  try {
    return new URL(normalizeUrl(url)).host;
  } catch {
    return '';
  }
}

export function makeId(): string {
  return `inst-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}