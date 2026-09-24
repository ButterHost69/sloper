'use client';

import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useInstanceData } from '@/hooks/use-instance-data';
import { api } from '@/lib/api';
import type { Health } from '@/lib/types';

interface HealthContextValue {
  health: Health | null;
  error: Error | null;
  refresh: () => void;
  refreshing: boolean;
  lastSuccessfulAt: number | null;
}

const HealthContext = createContext<HealthContextValue | null>(null);

/**
 * Single poller for `/api/health` shared by the shell and every page, so the
 * instance is not probed twice per cycle.
 */
export function HealthProvider({ children }: { children: ReactNode }) {
  const { data, error, refresh, refreshing, lastSuccessfulAt } = useInstanceData(
    (base) => api.health(base),
    15000,
  );

  const value = useMemo(
    () => ({ health: data, error, refresh, refreshing, lastSuccessfulAt }),
    [data, error, refresh, refreshing, lastSuccessfulAt],
  );

  return <HealthContext.Provider value={value}>{children}</HealthContext.Provider>;
}

export function useHealth(): HealthContextValue {
  const ctx = useContext(HealthContext);
  if (!ctx) throw new Error('useHealth must be used within HealthProvider');
  return ctx;
}
