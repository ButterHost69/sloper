export interface Instance {
  id: string;
  name: string;
  url: string;
  token?: string;
}

export interface Health {
  status: string;
  repo: string;
  version: string;
  go: string;
  uptime_s: number;
  time: string;
}

export interface RepoInfo {
  name: string;
  url: string;
  db_size: number;
  agent: {
    model: string;
    provider: string;
    bot_user: string;
    repo_env: string;
  };
}

export interface Summary {
  repo: string;
  issues: {
    total: number;
    open: number;
    closed: number;
    by_stage: Record<string, number>;
    failed: number;
    merged: number;
    in_flight: number;
  };
  runs: {
    total: number;
    by_status: Record<string, number>;
    running: number;
    failed: number;
    completed: number;
    interrupted: number;
  };
  pulls: {
    total: number;
    open: number;
    merged: number;
    closed: number;
  };
  events: {
    total: number;
    by_type: Record<string, number>;
  };
}

export interface Issue {
  number: number;
  title: string;
  state: string;
  url: string;
  author: string;
  updated_at: string;
  labels: string[];
  is_pull_request: boolean;
  stage: string;
  branch_name: string;
  pr_number: number;
  last_comment_id: number;
  review_iterations: number;
  created_at: string;
  first_seen_at: string;
  updated_at_local: string;
}

export interface CommentRecord {
  id: number;
  issue_number: number;
  author: string;
  body: string;
  created_at: string;
  processed: boolean;
  in_reply_to_id: number;
  replied_by_bot: boolean;
}

export interface RunRecord {
  id: number;
  issue_number: number;
  stage: string;
  status: string;
  checkpoint_json: string;
  agent_output: string;
  agent_thinking: string;
  shell_log: string;
  started_at: string;
  ended_at: string;
  error_message: string;
}

export interface EventRecord {
  id: number;
  issue_number: number;
  pr_number: number;
  event_type: string;
  stage: string;
  message: string;
  context: Record<string, unknown>;
  created_at: string;
}

export interface ActivityBucket {
  time: string;
  count: number;
}

export interface PullRecord {
  number: number;
  issue_number: number;
  title: string;
  head_sha: string;
  base_sha: string;
  state: string;
  raw_state: string;
  url: string;
  updated_at: string;
  merged_at: string;
  review_state: string;
  last_review_at: string;
}

export interface Spec {
  summary: string;
  files_to_change: string[];
  implementation_plan: string;
}

export interface IssueDetail {
  issue: Issue;
  spec: Spec | null;
  comments: CommentRecord[] | null;
  runs: RunRecord[] | null;
  events: EventRecord[] | null;
  pr: PullRecord | null;
}

export type WorktreeKind = 'work' | 'review' | 'fix' | 'unknown';

/** The issue a worktree belongs to, read from the database. */
export interface WorktreeIssue {
  number: number;
  title: string;
  state: string;
  stage: string;
  branch_name: string;
  pr_number: number;
  pr_state: string;
  pr_review_state: string;
  updated_at: string;
  run_stage: string;
  run_status: string;
}

/**
 * One worktree. `rel_path` is relative to the instance's worktree directory,
 * never absolute, so a remote console learns nothing about the host layout.
 */
export interface Worktree {
  rel_path: string;
  kind: WorktreeKind;
  live: boolean;
  mod_time: string;
  issue: WorktreeIssue | null;
}

export interface WorktreesResponse {
  base_dir: string;
  /** Directories on disk right now — an agent is working in these. */
  live: Worktree[];
  /** Issues that still have unmerged work, live or not. */
  open: Worktree[];
}

// ─── Agent sessions (pi transcripts) ─────────────────────────────────

export interface SessionToolRef {
  id: string;
  name: string;
  preview?: string;
  status: 'pending' | 'ok' | 'error' | string;
}

/** What the newest entry says the worker is doing right now. */
export interface SessionCurrent {
  kind: 'starting' | 'waiting' | 'thinking' | 'tool' | 'error' | string;
  text?: string;
  since?: string;
  tools?: SessionToolRef[];
}

export interface SessionSummary {
  title?: string;
  model?: string;
  provider?: string;
  messages: number;
  user_messages: number;
  assistant_messages: number;
  tool_results: number;
  tool_calls: number;
  tool_errors: number;
  thinking_chars: number;
  output_chars: number;
  total_tokens: number;
  cost_usd: number;
  compactions: number;
  started_at?: string;
  last_entry_at?: string;
  last_text?: string;
  current?: SessionCurrent | null;
  tools?: SessionToolRef[] | null;
}

export interface SessionRun {
  id: number;
  stage: string;
  status: string;
  started_at: string;
  ended_at: string;
  error_message?: string;
  duration_ms?: number;
}

export interface SessionPR {
  number: number;
  state: string;
  review_state?: string;
  url?: string;
}

export interface Session {
  id: string;
  file: string;
  issue_number: number;
  stage?: string;
  fix_iteration?: number;
  size_bytes: number;
  modified_at: string;
  started_at?: string;
  last_entry_at?: string;
  age_seconds: number;
  live: boolean;
  summary: SessionSummary;
  run?: SessionRun | null;
  pr?: SessionPR | null;
  issue_title?: string;
  issue_state?: string;
  issue_stage?: string;
  branch_name?: string;
}

export interface SessionGroup {
  issue_number: number;
  title?: string;
  state?: string;
  stage?: string;
  branch_name?: string;
  pr?: SessionPR | null;
  sessions: Session[];
  live_count: number;
  updated_at?: string;
  unscoped?: boolean;
}

export interface SessionsResponse {
  dir: string;
  groups: SessionGroup[];
  sessions: Session[];
  count: number;
  live_count: number;
  generated_at: string;
}

export interface SessionUsage {
  input?: number;
  output?: number;
  cacheRead?: number;
  cacheWrite?: number;
  reasoning?: number;
  totalTokens?: number;
  cost?: { input?: number; output?: number; cacheRead?: number; cacheWrite?: number; total?: number };
}

export interface SessionContentPart {
  type: string;
  text?: string;
  thinking?: string;
  id?: string;
  name?: string;
  arguments?: Record<string, unknown>;
  mimeType?: string;
  data?: string;
}

export interface SessionMessage {
  role: string;
  content?: SessionContentPart[];
  provider?: string;
  model?: string;
  usage?: SessionUsage;
  stopReason?: string;
  errorMessage?: string;
  toolCallId?: string;
  toolName?: string;
  isError?: boolean;
  timestamp?: number;
  /** bashExecution messages carry the command and its captured output. */
  command?: string;
  output?: string;
  exitCode?: number;
  cancelled?: boolean;
  truncated?: boolean;
}

/** One JSONL line of a pi session file, normalized by the API. */
export interface SessionEntry {
  type: string;
  id?: string;
  parent_id?: string;
  timestamp?: string;
  message?: SessionMessage;
  provider?: string;
  model_id?: string;
  level?: string;
  summary?: string;
  name?: string;
  label?: string;
  target_id?: string;
  custom_type?: string;
  from_id?: string;
  raw?: unknown;
}

export interface SessionDetail {
  session: Session;
  entries: SessionEntry[];
  start_offset: number;
  next_offset: number;
  has_more_before: boolean;
  size_bytes: number;
  modified_at: string;
  live: boolean;
  skipped?: number;
}

export interface SessionEventsResponse {
  entries: SessionEntry[];
  start_offset: number;
  next_offset: number;
  reset: boolean;
  has_more: boolean;
  has_more_before: boolean;
  size_bytes: number;
  modified_at: string;
  live: boolean;
  skipped?: number;
}
export interface ControllerHealth {
  status: string;
  version: string;
  uptime_s: number;
  time: string;
  docker: string;
  docker_error?: string;
  image: string;
  repos: number;
  repos_by_state?: Record<string, number>;
}

/**
 * One repo the controller runs a container for. `api_url` is that container's
 * sloper-web base URL — the console registers it as an instance.
 */
export interface AttachedRepo {
  id: number;
  link: string;
  name: string;
  clone_url: string;
  container_name: string;
  container_id: string;
  image: string;
  host_port: number;
  api_url: string;
  token: string;
  desired_state: 'running' | 'stopped' | string;
  status: string;
  status_message: string;
  container_state: string;
  restart_count: number;
  attached_at: string;
  updated_at: string;
}
