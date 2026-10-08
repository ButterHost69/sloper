'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import clsx from 'clsx';
import {
  Boxes,
  ExternalLink,
  GitBranch,
  Play,
  Plus,
  RefreshCw,
  RotateCcw,
  ScrollText,
  Server,
  Square,
  Trash2,
  XCircle,
} from 'lucide-react';
import { useInstances } from '@/components/instance-context';
import { EmptyState, Panel, PulseDot, SectionHeader, Spinner } from '@/components/ui';
import { PageHeader } from '@/components/page-header';
import { controllerApi, loadController, saveController, type ControllerConfig } from '@/lib/controller';
import { normalizeUrl } from '@/lib/instances';
import { timeAgo } from '@/lib/format';
import type { AttachedRepo, ControllerHealth } from '@/lib/types';

const POLL_MS = 5000;

/** Status -> the one color and label the card uses. */
function statusMeta(repo: AttachedRepo): { label: string; color: string; tone: string } {
  switch (repo.status) {
    case 'running':
      return { label: 'running', color: '#34d399', tone: 'text-emerald border-emerald/30 bg-emerald/10' };
    case 'starting':
      return { label: 'starting', color: '#38bdf8', tone: 'text-info border-info/30 bg-info/10' };
    case 'pending':
      return { label: 'pending', color: '#38bdf8', tone: 'text-info border-info/30 bg-info/10' };
    case 'stopped':
      return { label: 'stopped', color: '#94a3b8', tone: 'text-ink-faint border-edge bg-panel-2' };
    default:
      return { label: repo.status || 'error', color: '#ff6b6b', tone: 'text-danger border-danger/30 bg-danger/10' };
  }
}

function RepoCard({
  repo,
  busy,
  onOpen,
  onRestart,
  onToggle,
  onLogs,
  onDetach,
}: {
  repo: AttachedRepo;
  busy: boolean;
  onOpen: () => void;
  onRestart: () => void;
  onToggle: () => void;
  onLogs: () => void;
  onDetach: (purge: boolean) => void;
}) {
  const [confirming, setConfirming] = useState(false);
  const [purge, setPurge] = useState(false);
  const cancelRef = useRef<HTMLButtonElement>(null);
  const meta = statusMeta(repo);
  const running = repo.status === 'running';

  useEffect(() => {
    if (confirming) cancelRef.current?.focus();
  }, [confirming]);

  return (
    <div className="panel panel-hover p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-start sm:justify-between">
        <div className="flex min-w-0 flex-1 items-start gap-3">
          <div
            className={clsx('flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border', meta.tone)}
            aria-hidden="true"
          >
            {repo.status === 'pending' || repo.status === 'starting' ? (
              <PulseDot color={meta.color} />
            ) : repo.status === 'error' ? (
              <XCircle size={17} />
            ) : (
              <GitBranch size={17} />
            )}
          </div>

          <div className="min-w-0">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <p className="truncate font-semibold text-ink">{repo.name}</p>
              <span className={clsx('chip !text-[0.58rem]', meta.tone)}>{meta.label}</span>
              {repo.desired_state === 'stopped' && repo.status !== 'stopped' && (
                <span className="chip !text-[0.58rem]">stopping…</span>
              )}
            </div>
            <p className="mt-0.5 truncate font-mono text-xs text-ink-faint">{repo.link}</p>
            {repo.status_message && (
              <p
                className={clsx(
                  'mt-1 text-[11px] leading-4',
                  repo.status === 'error' ? 'text-danger' : 'text-ink-faint',
                )}
              >
                {repo.status_message}
              </p>
            )}
          </div>
        </div>

        <div className="flex shrink-0 flex-wrap items-center gap-1 sm:justify-end">
          <button
            type="button"
            className="btn !min-h-8 !px-2.5 !py-1.5 text-[11px]"
            onClick={onOpen}
            disabled={!repo.api_url}
            title="Open this repo's instance in the console"
          >
            <ExternalLink size={13} /> Open
          </button>
          <button
            type="button"
            className="btn btn-ghost !p-2"
            onClick={onRestart}
            disabled={busy}
            title="Restart container"
            aria-label={`Restart ${repo.name}`}
          >
            {busy ? <Spinner className="!h-3.5 !w-3.5" /> : <RotateCcw size={14} />}
          </button>
          <button
            type="button"
            className="btn btn-ghost !p-2"
            onClick={onToggle}
            disabled={busy}
            title={repo.desired_state === 'stopped' ? 'Start container' : 'Stop container'}
            aria-label={repo.desired_state === 'stopped' ? `Start ${repo.name}` : `Stop ${repo.name}`}
          >
            {repo.desired_state === 'stopped' ? <Play size={14} /> : <Square size={14} />}
          </button>
          <button
            type="button"
            className="btn btn-ghost !p-2"
            onClick={onLogs}
            title="Container logs"
            aria-label={`Logs for ${repo.name}`}
          >
            <ScrollText size={14} />
          </button>

          {confirming ? (
            <div
              className="flex flex-wrap items-center gap-2 rounded-lg border border-danger/30 bg-danger/[0.06] p-1.5"
              role="alert"
            >
              <span className="px-1 text-[11px] text-ink-dim">Remove container?</span>
              <label className="flex items-center gap-1.5 text-[11px] text-ink-dim">
                <input
                  type="checkbox"
                  className="accent-danger"
                  checked={purge}
                  onChange={(e) => setPurge(e.target.checked)}
                />
                delete its data
              </label>
              <button
                type="button"
                ref={cancelRef}
                className="btn btn-ghost !min-h-8 !px-2 !py-1 text-[11px]"
                onClick={() => setConfirming(false)}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn !min-h-8 !border-danger/40 !bg-danger/10 !px-2 !py-1 text-[11px] !text-danger"
                onClick={() => {
                  setConfirming(false);
                  onDetach(purge);
                  setPurge(false);
                }}
                aria-label={`Confirm removing ${repo.name}`}
              >
                Remove
              </button>
            </div>
          ) : (
            <button
              type="button"
              className="btn btn-ghost !p-2"
              onClick={() => setConfirming(true)}
              disabled={busy}
              title="Detach repo and remove its container"
              aria-label={`Detach ${repo.name}`}
            >
              <Trash2 size={14} className="text-danger" />
            </button>
          )}
        </div>
      </div>

      <div className="mt-4 flex flex-wrap items-center gap-x-4 gap-y-1.5 border-t border-edge pt-3 text-xs text-ink-dim">
        {repo.api_url && (
          <a
            href={repo.api_url}
            target="_blank"
            rel="noreferrer"
            className="hit-target flex items-center gap-1.5 font-mono text-accent hover:underline"
          >
            <Server size={12} /> {repo.api_url}
          </a>
        )}
        <span>
          container: <span className="font-mono text-ink-faint">{repo.container_name || '—'}</span>
        </span>
        <span>
          image: <span className="font-mono text-ink-faint">{repo.image}</span>
        </span>
        {repo.restart_count > 0 && (
          <span className="text-warn">restarted {repo.restart_count}×</span>
        )}
        <span className="ml-auto text-ink-faint">attached {timeAgo(repo.attached_at)}</span>
      </div>
    </div>
  );
}

export default function ReposPage() {
  const router = useRouter();
  const { instances, addInstance, updateInstance, setActive } = useInstances();

  const [cfg, setCfg] = useState<ControllerConfig>({ url: '' });
  const [urlDraft, setUrlDraft] = useState('');
  const [tokenDraft, setTokenDraft] = useState('');
  const [showSettings, setShowSettings] = useState(false);

  const [repos, setRepos] = useState<AttachedRepo[]>([]);
  const [health, setHealth] = useState<ControllerHealth | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [busyId, setBusyId] = useState<number | null>(null);

  const [showForm, setShowForm] = useState(false);
  const [link, setLink] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [logs, setLogs] = useState<{ repo: AttachedRepo; text: string } | null>(null);

  // The controller config is per-browser, so it is read after mount.
  useEffect(() => {
    const loaded = loadController();
    setCfg(loaded);
    setUrlDraft(loaded.url);
    setTokenDraft(loaded.token ?? '');
  }, []);

  const refresh = useCallback(
    async (config: ControllerConfig) => {
      if (!config.url) return;
      try {
        const [nextHealth, list] = await Promise.all([
          controllerApi.health(config),
          controllerApi.listRepos(config),
        ]);
        setHealth(nextHealth);
        setRepos(list.repos ?? []);
        setError(null);
      } catch (err) {
        setHealth(null);
        setError(err instanceof Error ? err.message : 'Controller unreachable');
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  useEffect(() => {
    if (!cfg.url) return;
    void refresh(cfg);
    const timer = setInterval(() => void refresh(cfg), POLL_MS);
    return () => clearInterval(timer);
  }, [cfg, refresh]);

  /**
   * Registers a repo's dashboard API as a console instance. Attaching does not
   * switch the active instance: the container may still be starting.
   */
  const ensureInstance = useCallback(
    (repo: AttachedRepo) => {
      if (!repo.api_url) return null;
      const existing = instances.find((i) => normalizeUrl(i.url) === normalizeUrl(repo.api_url));
      if (existing) {
        if (repo.token && existing.token !== repo.token) {
          updateInstance(existing.id, { token: repo.token });
        }
        return existing.id;
      }
      const created = addInstance(repo.name, repo.api_url, repo.token || undefined, {
        activate: false,
      });
      return created.id;
    },
    [addInstance, instances, updateInstance],
  );

  const attach = async (e: React.FormEvent) => {
    e.preventDefault();
    const value = link.trim();
    if (!value) {
      setFormError('Paste a repo link, for example owner/repo.');
      return;
    }
    setSubmitting(true);
    setFormError(null);
    try {
      const repo = await controllerApi.attachRepo(cfg, value);
      ensureInstance(repo);
      setNotice(`${repo.name} attached — its container is starting.`);
      setLink('');
      setShowForm(false);
      await refresh(cfg);
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Attach failed');
    } finally {
      setSubmitting(false);
    }
  };

  const withBusy = async (id: number, action: () => Promise<unknown>) => {
    setBusyId(id);
    setError(null);
    try {
      await action();
      await refresh(cfg);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setBusyId(null);
    }
  };

  const openRepo = (repo: AttachedRepo) => {
    const id = ensureInstance(repo);
    if (!id) return;
    setActive(id);
    router.push('/');
  };

  const showLogs = async (repo: AttachedRepo) => {
    try {
      const result = await controllerApi.logs(cfg, repo.id);
      setLogs({ repo, text: result.logs || '(no output yet)' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not read logs');
    }
  };

  const saveSettings = (e: React.FormEvent) => {
    e.preventDefault();
    const next: ControllerConfig = { url: urlDraft.trim(), token: tokenDraft.trim() || undefined };
    if (!next.url) {
      setError('Enter the controller URL, for example http://localhost:9090');
      return;
    }
    saveController(next);
    setCfg({ ...next, url: normalizeUrl(next.url) });
    setShowSettings(false);
    setLoading(true);
    setError(null);
  };

  const unreachable = !loading && !!error;
  const showSetup = unreachable || showSettings;

  return (
    <div className="space-y-5">
      <PageHeader
        eyebrow="Repositories"
        title="Repositories"
        subtitle="Attach a GitHub repo and Sloper starts a dedicated docker instance for it — clone, agent and pipeline included."
        actions={
          <>
            <button
              type="button"
              className="btn btn-ghost"
              onClick={() => setShowSettings((v) => !v)}
              aria-expanded={showSetup}
            >
              <Server size={14} /> Controller
            </button>
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => {
                setShowForm((v) => !v);
                setFormError(null);
              }}
            >
              <Plus size={14} /> {showForm ? 'Close' : 'New repo'}
            </button>
          </>
        }
      />

      {showForm && (
        <Panel>
          <form onSubmit={attach} className="space-y-3">
            <div>
              <label className="label" htmlFor="repo-link">
                Repo link
              </label>
              <input
                id="repo-link"
                className="input mono"
                placeholder="owner/repo or https://github.com/owner/repo"
                value={link}
                onChange={(e) => setLink(e.target.value)}
                aria-invalid={Boolean(formError)}
                aria-describedby={formError ? 'repo-link-error' : 'repo-link-hint'}
                autoFocus
              />
              <p id="repo-link-hint" className="mt-1.5 text-[11px] leading-4 text-ink-faint">
                The controller hands the repo to its own container on the next port in the range,
                then clones, specs and works on it. Credentials come from the controller&apos;s
                environment.
              </p>
              {formError && (
                <p id="repo-link-error" className="mt-1.5 text-[11px] text-danger" role="alert">
                  {formError}
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              <button type="submit" className="btn btn-primary" disabled={submitting}>
                {submitting ? <Spinner className="!h-3.5 !w-3.5" /> : <Plus size={14} />} Attach repo
              </button>
              <button
                type="button"
                className="btn btn-ghost"
                onClick={() => {
                  setShowForm(false);
                  setFormError(null);
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        </Panel>
      )}

      {notice && (
        <div
          className="panel flex items-center gap-2 border-emerald/30 bg-emerald/[0.06] px-4 py-2.5 text-xs text-emerald"
          role="status"
        >
          <Boxes size={14} /> {notice}
          <button
            type="button"
            className="ml-auto text-accent hover:underline"
            onClick={() => setNotice(null)}
          >
            dismiss
          </button>
        </div>
      )}

      {showSetup && (
        <Panel title="Controller">
          <form onSubmit={saveSettings} className="space-y-3">
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-[2fr_1fr]">
              <div>
                <label className="label" htmlFor="controller-url">
                  Controller URL
                </label>
                <input
                  id="controller-url"
                  className="input mono"
                  placeholder="http://localhost:9090"
                  value={urlDraft}
                  onChange={(e) => setUrlDraft(e.target.value)}
                />
              </div>
              <div>
                <label className="label" htmlFor="controller-token">
                  Bearer token (optional)
                </label>
                <input
                  id="controller-token"
                  type="password"
                  className="input mono"
                  placeholder="SLOPER_CONTROLLER_TOKEN"
                  value={tokenDraft}
                  onChange={(e) => setTokenDraft(e.target.value)}
                />
              </div>
            </div>
            {error && (
              <p className="text-[11px] text-danger" role="alert">
                {error}
              </p>
            )}
            <p className="text-[11px] leading-4 text-ink-faint">
              The controller is the backend that runs containers. Start it with{' '}
              <code className="font-mono text-ink-dim">make run-controller</code> on the docker host,
              set <code className="font-mono text-ink-dim">GH_TOKEN</code> and{' '}
              <code className="font-mono text-ink-dim">AGENT_MODEL</code> for it, then point this page
              at its address.
            </p>
            <button type="submit" className="btn btn-primary">
              Save and connect
            </button>
          </form>
        </Panel>
      )}

      <div>
        <SectionHeader
          title="Attached repos"
          sub={
            health
              ? `controller ${health.version} · image ${health.image} · docker ${health.docker}`
              : 'one container per attached repo, kept alive by the controller'
          }
          action={
            <button
              type="button"
              className="btn btn-ghost !min-h-8 !px-2.5 !py-1.5 text-[11px]"
              onClick={() => void refresh(cfg)}
            >
              <RefreshCw size={13} /> Refresh
            </button>
          }
        />

        <div className="space-y-3">
          {loading && repos.length === 0 ? (
            <Panel>
              <p className="flex items-center gap-2 text-xs text-ink-faint">
                <PulseDot color="#38bdf8" /> asking the controller…
              </p>
            </Panel>
          ) : repos.length === 0 ? (
            <Panel>
              <EmptyState
                icon={<Boxes size={19} />}
                title="No repos attached yet"
                hint="Paste a repo link and Sloper starts a container for it: clone, worktrees, agent runs and its own dashboard API."
                action={
                  <button
                    type="button"
                    className="btn btn-primary mt-1"
                    onClick={() => setShowForm(true)}
                  >
                    <Plus size={14} /> New repo
                  </button>
                }
              />
            </Panel>
          ) : (
            repos.map((repo) => (
              <RepoCard
                key={repo.id}
                repo={repo}
                busy={busyId === repo.id}
                onOpen={() => openRepo(repo)}
                onRestart={() => void withBusy(repo.id, () => controllerApi.restartRepo(cfg, repo.id))}
                onToggle={() =>
                  void withBusy(repo.id, () =>
                    controllerApi.setDesiredState(
                      cfg,
                      repo.id,
                      repo.desired_state === 'stopped' ? 'running' : 'stopped',
                    ),
                  )
                }
                onLogs={() => void showLogs(repo)}
                onDetach={(purge) =>
                  void withBusy(repo.id, () => controllerApi.detachRepo(cfg, repo.id, purge))
                }
              />
            ))
          )}
        </div>
      </div>

      {logs && (
        <Panel
          title={`${logs.repo.name} · container logs`}
          action={
            <button
              type="button"
              className="btn btn-ghost !min-h-8 !px-2.5 !py-1.5 text-[11px]"
              onClick={() => setLogs(null)}
            >
              Close
            </button>
          }
        >
          <pre className="surface-grid max-h-96 overflow-auto rounded-lg border border-edge bg-panel-2 p-3 font-mono text-[11px] leading-5 text-ink-dim">
            {logs.text}
          </pre>
        </Panel>
      )}
    </div>
  );
}
