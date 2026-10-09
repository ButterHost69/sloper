-- 0004_repos.sql — Repos the controller runs a dedicated docker instance for.
--
-- One row per attached repo. The row is the desired state (the repo should have
-- a container named container_name running); the actual container state is
-- written back on every reconcile so the console can show what is really up.

CREATE TABLE IF NOT EXISTS repos (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    link             TEXT NOT NULL UNIQUE,           -- owner/repo, the attach key
    name             TEXT NOT NULL DEFAULT '',       -- short repo name for display
    clone_url        TEXT NOT NULL DEFAULT '',       -- https clone URL
    container_name   TEXT NOT NULL DEFAULT '',       -- docker container for this repo
    container_id     TEXT NOT NULL DEFAULT '',       -- docker container id (short)
    image            TEXT NOT NULL DEFAULT '',       -- image the container runs
    host_port        INTEGER NOT NULL DEFAULT 0,     -- published sloper-web port
    api_token        TEXT NOT NULL DEFAULT '',       -- SLOPER_WEB_TOKEN of the instance
    desired_state    TEXT NOT NULL DEFAULT 'running',-- running|stopped
    status           TEXT NOT NULL DEFAULT 'pending',-- pending|starting|running|stopped|error
    status_message   TEXT NOT NULL DEFAULT '',       -- last reconcile error, if any
    container_state  TEXT NOT NULL DEFAULT '',       -- raw docker state when known
    restart_count    INTEGER NOT NULL DEFAULT 0,     -- consecutive restarts of a container that keeps exiting
    attached_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_repos_status ON repos(status);
