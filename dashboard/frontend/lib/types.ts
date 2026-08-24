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