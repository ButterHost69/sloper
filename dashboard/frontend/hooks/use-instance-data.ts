'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { ApiError } from '@/lib/api';
import { useInstances } from '@/components/instance-context';

export interface DataState<T> {
  data: T | null;
  error: ApiError | Error | null;
  loading: boolean;
  refreshing: boolean;
  refresh: () => void;
}

/**
 * Fetches from the active sloper instance, re-polling every `intervalMs`.
 * Falls back to the previous data during re-fetch to avoid UI flicker.
 *
 * The fetcher is kept in a ref so its (typically inline) identity does not
 * retrigger the polling effects — only a change of instance URL does. Pass
 * extra values in `deps` (e.g. search/filter state) to force a reload when
 * they change.
 */
export function useInstanceData<T>(
  fetcher: (base: string) => Promise<T>,
  intervalMs = 10000,
  enabled = true,
  deps: unknown[] = [],
): DataState<T> {
  const { active } = useInstances();
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const baseRef = useRef<string | null>(null);
  const fetcherRef = useRef(fetcher);

  // Always point at the latest fetcher without re-running effects.
  useEffect(() => {
    fetcherRef.current = fetcher;
  });

  const url = active?.url ?? null;

  const load = useCallback(
    async (isRefresh: boolean) => {
      if (!url) {
        setData(null);
        setLoading(false);
        return;
      }
      if (!isRefresh) setLoading(true);
      else setRefreshing(true);
      try {
        const result = await fetcherRef.current(url);
        setData(result);
        setError(null);
        baseRef.current = url;
      } catch (err) {
        setError(err instanceof ApiError ? err : (err as Error));
        if (baseRef.current !== url) {
          // instance switched — drop stale data
          setData(null);
        }
      } finally {
        setLoading(false);
        setRefreshing(false);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [url, ...deps],
  );

  useEffect(() => {
    if (!enabled) return;
    void load(false);
  }, [load, enabled]);

  useEffect(() => {
    if (!enabled || !intervalMs) return;
    timerRef.current = setInterval(() => void load(true), intervalMs);
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [load, intervalMs, enabled]);

  const refresh = useCallback(() => void load(true), [load]);

  return { data, error, loading, refreshing, refresh };
}