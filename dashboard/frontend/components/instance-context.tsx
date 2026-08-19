'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import type { Instance } from '@/lib/types';
import {
  ACTIVE_KEY,
  loadActiveInstanceId,
  loadInstances,
  makeId,
  normalizeUrl,
  saveActiveInstanceId,
  saveInstances,
} from '@/lib/instances';

interface InstanceContextValue {
  instances: Instance[];
  active: Instance | null;
  setActive: (id: string) => void;
  addInstance: (name: string, url: string, token?: string) => Instance;
  updateInstance: (id: string, patch: Partial<Pick<Instance, 'name' | 'url' | 'token'>>) => void;
  removeInstance: (id: string) => void;
}

const InstanceContext = createContext<InstanceContextValue | null>(null);

export function InstanceProvider({ children }: { children: ReactNode }) {
  const [instances, setInstances] = useState<Instance[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  useEffect(() => {
    const list = loadInstances();
    setInstances(list);
    const saved = loadActiveInstanceId();
    if (saved && list.some((i) => i.id === saved)) {
      setActiveId(saved);
    } else {
      setActiveId(list[0]?.id ?? null);
    }
  }, []);

  const persist = useCallback((list: Instance[]) => {
    setInstances(list);
    saveInstances(list);
  }, []);

  const setActive = useCallback((id: string) => {
    setActiveId(id);
    saveActiveInstanceId(id);
  }, []);

  const addInstance = useCallback(
    (name: string, url: string, token?: string) => {
      const inst: Instance = { id: makeId(), name, url: normalizeUrl(url), token: token || undefined };
      persist([...instances, inst]);
      setActiveId(inst.id);
      saveActiveInstanceId(inst.id);
      return inst;
    },
    [instances, persist],
  );

  const updateInstance = useCallback(
    (id: string, patch: Partial<Pick<Instance, 'name' | 'url' | 'token'>>) => {
      persist(
        instances.map((i) =>
          i.id === id
            ? {
                ...i,
                ...patch,
                url: patch.url ? normalizeUrl(patch.url) : i.url,
                token: patch.token === undefined ? i.token : patch.token || undefined,
              }
            : i,
        ),
      );
    },
    [instances, persist],
  );

  const removeInstance = useCallback(
    (id: string) => {
      const next = instances.filter((i) => i.id !== id);
      persist(next.length ? next : []);
      if (activeId === id) {
        const fallback = next[0]?.id ?? null;
        setActiveId(fallback);
        if (fallback) saveActiveInstanceId(fallback);
        else localStorage.removeItem(ACTIVE_KEY);
      }
    },
    [instances, persist, activeId],
  );

  const active = useMemo(
    () => instances.find((i) => i.id === activeId) ?? null,
    [instances, activeId],
  );

  const value = useMemo(
    () => ({ instances, active, setActive, addInstance, updateInstance, removeInstance }),
    [instances, active, setActive, addInstance, updateInstance, removeInstance],
  );

  return <InstanceContext.Provider value={value}>{children}</InstanceContext.Provider>;
}

export function useInstances(): InstanceContextValue {
  const ctx = useContext(InstanceContext);
  if (!ctx) throw new Error('useInstances must be used within InstanceProvider');
  return ctx;
}