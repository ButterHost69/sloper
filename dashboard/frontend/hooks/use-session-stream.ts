'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { ApiError, api } from '@/lib/api';
import { useInstances } from '@/components/instance-context';
import type { Session, SessionEntry } from '@/lib/types';

/** Poll cadence while the worker is writing, and once it has gone quiet. */
const LIVE_POLL_MS = 1500;
const IDLE_POLL_MS = 6000;

export interface SessionStream {
  session: Session | null;
  entries: SessionEntry[];
  live: boolean;
  modifiedAt: string;
  loading: boolean;
  /** The session file was deleted (approve, abort, merge) while we watched. */
  gone: boolean;
  error: Error | null;
  hasMoreBefore: boolean;
  loadingEarlier: boolean;
  loadEarlier: () => void;
  reload: () => void;
}

/**
 * Streams one pi session file: an initial window, then incremental reads from
 * the byte offset the server hands back. Polling is recursive rather than an
 * interval so a slow response never stacks up requests.
 */
export function useSessionStream(id: string, pageSize = 400): SessionStream {
  const { active } = useInstances();
  const url = active?.url ?? null;

  const [session, setSession] = useState<Session | null>(null);
  const [entries, setEntries] = useState<SessionEntry[]>([]);
  const [live, setLive] = useState(false);
  const [modifiedAt, setModifiedAt] = useState('');
  const [loading, setLoading] = useState(true);
  const [gone, setGone] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [hasMoreBefore, setHasMoreBefore] = useState(false);
  const [loadingEarlier, setLoadingEarlier] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  const nextOffsetRef = useRef(0);
  const startOffsetRef = useRef(0);
  const earlierInFlightRef = useRef(false);
  const liveRef = useRef(false);
  liveRef.current = live;

  useEffect(() => {
    let cancelled = false;
    nextOffsetRef.current = 0;
    startOffsetRef.current = 0;
    setSession(null);
    setEntries([]);
    setLive(false);
    setGone(false);
    setError(null);
    setHasMoreBefore(false);
    setLoading(true);

    if (!url) {
      setLoading(false);
      return;
    }
    void (async () => {
      try {
        const detail = await api.session(url, id, pageSize);
        if (cancelled) return;
        setSession(detail.session);
        setEntries(detail.entries ?? []);
        nextOffsetRef.current = detail.next_offset;
        startOffsetRef.current = detail.start_offset;
        setHasMoreBefore(detail.has_more_before);
        setLive(detail.live);
        setModifiedAt(detail.modified_at);
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 404) setGone(true);
        else setError(err instanceof Error ? err : new Error('Could not load the session'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [url, id, pageSize, reloadKey]);

  useEffect(() => {
    if (!url || gone || loading) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const tick = async () => {
      try {
        const chunk = await api.sessionEvents(url, id, {
          offset: nextOffsetRef.current,
          limit: 500,
        });
        if (cancelled) return;
        if (chunk.reset) {
          // The file was rewritten (a fork or session switch): start over.
          setReloadKey((value) => value + 1);
          return;
        }
        if (chunk.entries?.length) {
          setEntries((previous) => [...previous, ...chunk.entries]);
        }
        nextOffsetRef.current = chunk.next_offset;
        setLive(chunk.live);
        if (chunk.modified_at) setModifiedAt(chunk.modified_at);
      } catch (err) {
        if (cancelled) return;
        if (err instanceof ApiError && err.status === 404) {
          setGone(true);
          return;
        }
        // Transient failures keep the transcript on screen and retry.
      } finally {
        if (!cancelled) {
          timer = setTimeout(tick, liveRef.current ? LIVE_POLL_MS : IDLE_POLL_MS);
        }
      }
    };

    timer = setTimeout(tick, LIVE_POLL_MS);
    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
    };
  }, [url, id, gone, loading, reloadKey]);

  const loadEarlier = useCallback(() => {
    if (!url || earlierInFlightRef.current || startOffsetRef.current <= 0) return;
    earlierInFlightRef.current = true;
    setLoadingEarlier(true);
    void (async () => {
      try {
        const chunk = await api.sessionEvents(url, id, {
          before: startOffsetRef.current,
          limit: pageSize,
        });
        if (chunk.entries?.length) {
          setEntries((previous) => [...(chunk.entries ?? []), ...previous]);
        }
        startOffsetRef.current = chunk.start_offset;
        setHasMoreBefore(chunk.has_more_before);
      } catch {
        // Keep whatever is already rendered; the button stays available.
      } finally {
        earlierInFlightRef.current = false;
        setLoadingEarlier(false);
      }
    })();
  }, [url, id, pageSize]);

  const reload = useCallback(() => setReloadKey((value) => value + 1), []);

  return {
    session,
    entries,
    live,
    modifiedAt,
    loading,
    gone,
    error,
    hasMoreBefore,
    loadingEarlier,
    loadEarlier,
    reload,
  };
}
