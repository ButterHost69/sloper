-- 0001_init.sql — Initial schema for sloper SQLite database

-- Issue cache + pipeline state
CREATE TABLE IF NOT EXISTS issues (
    number            INTEGER PRIMARY KEY,
    title             TEXT NOT NULL DEFAULT '',
    state             TEXT NOT NULL DEFAULT 'open',
    url               TEXT NOT NULL DEFAULT '',
    author            TEXT NOT NULL DEFAULT '',
    updated_at        TEXT NOT NULL DEFAULT '',       -- from GitHub, used for change detection
    labels            TEXT NOT NULL DEFAULT '[]',     -- JSON array of label names
    is_pull_request   INTEGER NOT NULL DEFAULT 0,
    stage             TEXT NOT NULL DEFAULT 'new',    -- new|spec-done|approved|work-done|review-done|merged|failed
    spec_json         TEXT,                           -- cached SpecResult JSON (preserved between SPEC and WORK)
    branch_name       TEXT,
    pr_number         INTEGER,
    last_comment_id   INTEGER NOT NULL DEFAULT 0,    -- highest processed comment ID
    review_iterations INTEGER NOT NULL DEFAULT 0,
    created_at        TEXT NOT NULL DEFAULT '',       -- issue creation time on GitHub
    first_seen_at     TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at_local  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Issue comment cache (for slash-command detection + re-triage)
CREATE TABLE IF NOT EXISTS issue_comments (
    id              INTEGER PRIMARY KEY,             -- GitHub comment ID
    issue_number    INTEGER NOT NULL,
    author          TEXT NOT NULL DEFAULT '',
    body            TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT '',
    processed       INTEGER NOT NULL DEFAULT 0,      -- 1 = sloper has acted on this comment
    FOREIGN KEY (issue_number) REFERENCES issues(number)
);

CREATE INDEX IF NOT EXISTS idx_issue_comments_issue ON issue_comments(issue_number);
CREATE INDEX IF NOT EXISTS idx_issue_comments_unprocessed ON issue_comments(issue_number, processed);

-- Pull request cache
CREATE TABLE IF NOT EXISTS pull_requests (
    number            INTEGER PRIMARY KEY,
    issue_number      INTEGER,
    title             TEXT NOT NULL DEFAULT '',
    head_sha          TEXT NOT NULL DEFAULT '',
    base_sha          TEXT NOT NULL DEFAULT '',
    state             TEXT NOT NULL DEFAULT 'open',
    url               TEXT NOT NULL DEFAULT '',
    updated_at        TEXT NOT NULL DEFAULT '',
    review_state      TEXT NOT NULL DEFAULT 'none',  -- none|pending|approved|changes_requested
    last_review_at    TEXT,
    created_at_local  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_issue ON pull_requests(issue_number);

-- Pipeline runs (crash recovery + output persistence)
CREATE TABLE IF NOT EXISTS runs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_number    INTEGER NOT NULL,
    stage           TEXT NOT NULL,                   -- spec|work|review|fix|merge
    status          TEXT NOT NULL DEFAULT 'running', -- running|completed|failed|interrupted
    checkpoint_json TEXT,                            -- stage-specific resume state
    agent_output    TEXT,                            -- collected pi text output
    agent_thinking  TEXT,                            -- collected thinking output
    shell_log       TEXT,                            -- summary of shell commands run
    started_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    ended_at        TEXT,
    error_message   TEXT,
    FOREIGN KEY (issue_number) REFERENCES issues(number)
);

CREATE INDEX IF NOT EXISTS idx_runs_issue ON runs(issue_number);
CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);

-- Event/audit log (for debugging + future dashboard)
CREATE TABLE IF NOT EXISTS event_logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_number    INTEGER,
    pr_number       INTEGER,
    event_type      TEXT NOT NULL,                   -- tick.started, spec.completed, work.started, etc.
    stage           TEXT,
    message         TEXT NOT NULL DEFAULT '',
    context_json    TEXT NOT NULL DEFAULT '{}',
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_event_logs_issue ON event_logs(issue_number);
CREATE INDEX IF NOT EXISTS idx_event_logs_type ON event_logs(event_type);
