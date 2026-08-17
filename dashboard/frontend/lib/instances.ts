import type { Instance } from './types';

const STORAGE_KEY = 'sloper.instances.v1';
const ACTIVE_KEY = 'sloper.activeInstance.v1';

const DEFAULT_INSTANCE: Instance = {
  id: 'local',
  name: 'Local Sloper',
  url: 'http://localhost:8080',
};

export function loadInstances(): Instance[] {
  if (typeof window === 'undefined') return [DEFAULT_INSTANCE];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [DEFAULT_INSTANCE];
    const parsed = JSON.parse(raw) as Instance[];
    if (!Array.isArray(parsed) || parsed.length === 0) return [DEFAULT_INSTANCE];
    return parsed;
  } catch {
    return [DEFAULT_INSTANCE];
  }
}

export function saveInstances(instances: Instance[]): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(instances));
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

export function normalizeUrl(url: string): string {
  return url.trim().replace(/\/+$/, '');
}

export function makeId(): string {
  return `inst-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}