'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { CheckCircle2, ExternalLink, KeyRound, Pencil, Plus, Server, Trash2, XCircle } from 'lucide-react';
import clsx from 'clsx';
import { useInstances } from '@/components/instance-context';
import { api } from '@/lib/api';
import { normalizeUrl } from '@/lib/instances';
import { timeAgo, formatBytes } from '@/lib/format';
import type { Health, RepoInfo } from '@/lib/types';
import { Panel, PulseDot, SectionHeader, EmptyState } from '@/components/ui';
import { PageHeader } from '@/components/page-header';

interface Probe {
  health: Health | null;
  repo: RepoInfo | null;
  loading: boolean;
  error: string | null;
  lastChecked: number;
}

function useProbe(url: string | undefined, intervalMs = 20000) {
  const requestIdRef = useRef(0);
  const [probe, setProbe] = useState<Probe>({
    health: null,
    repo: null,
    loading: true,
    error: null,
    lastChecked: 0,
  });

  const check = useCallback(async () => {
    const requestId = ++requestIdRef.current;
    if (!url) {
      setProbe((p) => ({ ...p, loading: false }));
      return;
    }
    setProbe((p) => ({ ...p, loading: true }));
    try {
      const [health, repo] = await Promise.all([api.health(url), api.repo(url)]);
      if (requestId !== requestIdRef.current) return;
      setProbe({ health, repo, loading: false, error: null, lastChecked: Date.now() });
    } catch (err) {
      if (requestId !== requestIdRef.current) return;
      setProbe({
        health: null,
        repo: null,
        loading: false,
        error: err instanceof Error ? err.message : 'Failed to connect',
        lastChecked: Date.now(),
      });
    }
  }, [url]);

  useEffect(() => {
    void check();
    const t = setInterval(check, intervalMs);
    return () => clearInterval(t);
  }, [check, intervalMs]);

  return { probe, check };
}

function InstanceCard({
  id,
  name,
  url,
  isActive,
  onActivate,
  onEdit,
  onRemove,
}: {
  id: string;
  name: string;
  url: string;
  isActive: boolean;
  onActivate: () => void;
  onEdit: () => void;
  onRemove: () => void;
}) {
  const [confirming, setConfirming] = useState(false);
  const cancelRemoveRef = useRef<HTMLButtonElement>(null);
  const { probe, check } = useProbe(url);

  useEffect(() => {
    if (confirming) cancelRemoveRef.current?.focus();
  }, [confirming]);
  const ok = !!probe.health;

  return (
    <div
      className={clsx(
        'panel panel-hover p-4',
        isActive && 'border-accent/40 shadow-[0_0_0_1px_rgba(52,211,153,0.25),0_8px_30px_rgba(0,0,0,0.35)]',
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 flex-1 items-center gap-3">
          <div
            className={clsx(
              'flex h-10 w-10 items-center justify-center rounded-lg border',
              ok ? 'border-emerald/30 bg-emerald/10 text-emerald' : 'border-danger/30 bg-danger/10 text-danger',
            )}
          >
            {probe.loading ? (
              <PulseDot color="#38bdf8" />
            ) : ok ? (
              <Server size={17} />
            ) : (
              <XCircle size={17} />
            )}
          </div>
          <div className="min-w-0">
            <div className="flex min-w-0 items-center gap-2">
              <p className="truncate font-semibold text-ink">{name}</p>
              {isActive && (
                <span className="chip !text-[0.58rem] !text-accent !border-accent/40 !bg-accent/10">
                  active
                </span>
              )}
            </div>
            <p className="mt-0.5 truncate font-mono text-xs text-ink-faint">{url}</p>
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap items-center justify-end gap-1">
          {!isActive && (
            <button type="button" className="btn !min-h-8 !px-2.5 !py-1.5 text-[11px]" onClick={onActivate}>
              Use instance
            </button>
          )}
          <button
            type="button"
            className="btn btn-ghost !p-2"
            onClick={onEdit}
            title="Edit"
            aria-label={`Edit ${name}`}
          >
            <Pencil size={14} />
          </button>
          {confirming ? (
            <div className="flex items-center gap-1 rounded-lg border border-danger/30 bg-danger/[0.06] p-1" role="alert">
              <span className="px-1 text-[11px] text-ink-dim">Remove connection?</span>
              <button
                type="button"
                ref={cancelRemoveRef}
                className="btn btn-ghost !min-h-8 !px-2 !py-1 text-[11px]"
                onClick={() => setConfirming(false)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn !min-h-8 !border-danger/40 !bg-danger/10 !px-2 !py-1 text-[11px] !text-danger"
                onClick={onRemove}
                aria-label={`Confirm removing ${name}`}
              >
                Remove
              </button>
            </div>
          ) : (
            <button
              type="button"
              className="btn btn-ghost !p-2"
              onClick={() => setConfirming(true)}
              title="Remove"
              aria-label={`Remove ${name}`}
            >
              <Trash2 size={14} className="text-danger" />
            </button>
          )}
        </div>
      </div>

      <div className="mt-4 border-t border-edge pt-3">
        {probe.loading ? (
          <p className="flex items-center gap-2 text-xs text-ink-faint">
            <PulseDot color="#38bdf8" /> probing…
          </p>
        ) : ok ? (
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-ink-dim">
            <span className="flex items-center gap-1.5 text-emerald">
              <CheckCircle2 size={13} /> healthy
            </span>
            <span>repo: <span className="font-semibold text-ink">{probe.health?.repo}</span></span>
            {probe.repo?.db_size ? (
              <span>db {formatBytes(probe.repo.db_size)}</span>
            ) : null}
            <span>up {timeAgo(new Date(Date.now() - (probe.health?.uptime_s ?? 0) * 1000).toISOString())}</span>
            {probe.health?.time && (
              <span className="text-ink-faint">{timeAgo(probe.health.time)}</span>
            )}
            {probe.repo?.agent?.model && (
              <span>
                model: <span className="font-mono text-info">{probe.repo.agent.model}</span>
              </span>
            )}
            <button
              onClick={() => void check()}
              className="hit-target ml-auto text-xs text-accent hover:underline"
            >
              re-check
            </button>
          </div>
        ) : (
          <div className="flex items-center gap-2 text-xs text-danger">
            <XCircle size={13} />
            <span className="truncate">{probe.error ?? 'Unreachable'}</span>
            <button onClick={() => void check()} className="ml-auto text-accent hover:underline">
              retry
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

export default function InstancesPage() {
  const { instances, active, setActive, addInstance, updateInstance, removeInstance } =
    useInstances();
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [token, setToken] = useState('');
  const [formError, setFormError] = useState<string | null>(null);

  const reset = () => {
    setShowForm(false);
    setEditing(null);
    setName('');
    setUrl('');
    setToken('');
    setFormError(null);
  };

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !url.trim()) {
      setFormError('Name and URL are required.');
      return;
    }
    let normalized: string;
    try {
      normalized = normalizeUrl(url.trim());
      const parsed = new URL(normalized);
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') throw new Error();
    } catch {
      setFormError('URL must be a valid http:// or https:// address.');
      return;
    }
    const cleanToken = token.trim() || null;
    if (editing) {
      updateInstance(editing, { name: name.trim(), url: normalized, token: cleanToken });
    } else {
      addInstance(name.trim(), normalized, cleanToken || undefined);
    }
    reset();
  };

  const startEdit = (id: string, n: string, u: string, t?: string) => {
    setEditing(id);
    setName(n);
    setUrl(u);
    setToken(t ?? '');
    setShowForm(true);
    setFormError(null);
  };

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Instances"
        title="Instances"
        subtitle="Point the console at one or more running sloper web servers."
        actions={
          <button
            className="btn btn-primary"
            onClick={() => {
              reset();
              setShowForm((v) => !v);
            }}
          >
            <Plus size={14} /> {showForm ? 'Close' : 'Add instance'}
          </button>
        }
      />

      {showForm && (
        <Panel>
          <form onSubmit={submit} className="space-y-3">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1fr_2fr]">
              <div>
                <label className="label" htmlFor="instance-name">Name</label>
                <input
                  id="instance-name"
                  className="input"
                  aria-invalid={Boolean(formError)}
                  aria-describedby={formError ? 'instance-form-error' : undefined}
                  placeholder="e.g. Staging"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoFocus
                />
              </div>
              <div>
                <label className="label" htmlFor="instance-url">Base URL</label>
                <input
                  id="instance-url"
                  className="input mono"
                  aria-invalid={Boolean(formError)}
                  aria-describedby={formError ? 'instance-form-error' : undefined}
                  placeholder="http://localhost:8080"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                />
              </div>
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <label className="label flex items-center gap-1.5" htmlFor="instance-token">
                  <KeyRound size={12} className="text-ink-faint" /> Bearer token (optional)
                </label>
                <input
                  id="instance-token"
                  type="password"
                  className="input mono"
                  placeholder="SLOPER_WEB_TOKEN if the server requires one"
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  autoComplete="off"
                />
              </div>
            </div>
            {formError && (
              <p id="instance-form-error" role="alert" className="flex items-center gap-1.5 text-xs text-danger">
                <XCircle size={13} /> {formError}
              </p>
            )}
            <div className="flex items-center gap-2">
              <button type="submit" className="btn btn-primary">
                {editing ? 'Save' : 'Add'}
              </button>
              <button type="button" className="btn" onClick={reset}>
                Cancel
              </button>
            </div>
          </form>
        </Panel>
      )}

      <div className="space-y-3">
        {instances.map((inst) => (
          <InstanceCard
            key={inst.id}
            id={inst.id}
            name={inst.name}
            url={inst.url}
            isActive={inst.id === active?.id}
            onActivate={() => setActive(inst.id)}
            onEdit={() => startEdit(inst.id, inst.name, inst.url, inst.token)}
            onRemove={() => removeInstance(inst.id)}
          />
        ))}
      </div>

      {instances.length === 0 && (
        <Panel bodyClassName="p-0">
          <EmptyState
            icon={<Server size={19} />}
            title="No instances configured"
            hint="Add your first Sloper web server to begin observing it."
            action={
              <button type="button" className="btn btn-primary mt-1" onClick={() => setShowForm(true)}>
                <Plus size={14} /> Add instance
              </button>
            }
          />
        </Panel>
      )}

      <div className="pt-4">
        <SectionHeader title="How to expose a sloper instance" sub="Run the built-in API server next to each sloper instance" />
        <Panel bodyClassName="p-5">
          <p className="text-sm text-ink-dim">
            Each sloper instance ships a read-only JSON API backed by its own SQLite database. Start it
            wherever sloper is running:
          </p>
          <pre className="mt-3 overflow-x-auto rounded-lg border border-edge bg-base p-4 font-mono text-xs leading-relaxed text-ink-dim">
{`# from a machine with access to the sloper database
SLOPER_DB_PATH=/path/to/sloper.sqlite \\
SLOPER_WEB_PORT=8080 \\
  ./dist/sloper-web

# defaults: ~/.sloper/sloper.sqlite on 127.0.0.1:8080
# env: SLOPER_WEB_PORT, SLOPER_WEB_ADDR, SLOPER_DB_PATH, SLOPER_REPO
# optional auth: SLOPER_WEB_TOKEN=secret — add the same token to the form above`}
          </pre>
          <div className="mt-4 flex items-start gap-2 rounded-lg border border-edge bg-panel-2 p-3 text-xs text-ink-dim">
            <ExternalLink size={14} className="mt-0.5 shrink-0 text-accent" />
            <p>
              Make sure the port is reachable from the browser — the dashboard calls the instance API
              directly from your browser (CORS is enabled on the server). For remote containers use
              <code className="mono text-info"> docker run -p 8080:8080</code> or a reverse proxy.
            </p>
          </div>
        </Panel>
      </div>
    </div>
  );
}