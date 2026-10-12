# Sloper Console

A light/dark observability dashboard for **one or more running sloper instances**,
built with Next.js (App Router) + Tailwind CSS + Recharts.

The console is a pure frontend. It talks directly (from your browser) to each
sloper instance's read-only JSON API, which is served by the small Go server in
[`app/web`](../../app/web).

## Architecture

```
┌──────────────────────┐     browser fetch (CORS enabled)      ┌──────────────────────┐
│  Next.js dashboard   │ ────────────────────────────────────▶ │  sloper instance #1  │
│  dashboard/frontend  │                                       │  sloper-web :8080    │
│                      │                                       │  └── sqlite db       │
│  • instance registry │                                       ├──────────────────────┤
│  • overview/issues   │                                       │  sloper instance #2  │
│  • PR/runs/events    │                                       │  sloper-web :8081    │
│  • pipeline DAGs     │                                       │  └── sqlite db       │
└──────────────────────┘                                       └──────────────────────┘
```

- The dashboard keeps its instance list (name + base URL) in `localStorage`.
- Each instance must run `sloper-web` (the API server) with its own database.
- Everything is read-only: the console observes, it never mutates GitHub or the DB.

## Visual language

The console uses a restrained Apple-inspired Liquid Glass system: solid pale neutral canvases, one blue action accent, layered translucent shell chrome with blur/saturation and light-catching edges, luminous hairlines, soft elevation, capsule controls, and strong ease-out motion. Glass is reserved for the sidebar, sticky header, menus, and temporary overlays; operational panels stay solid. It follows the token-first and accessibility guidance in the OpenDesign [glassmorphism design system](https://github.com/nexu-io/open-design/tree/main/design-systems/glassmorphism) and the precision/contrast rules in its [Apple design system](https://github.com/nexu-io/open-design/tree/main/design-systems/apple). Dark mode keeps the same semantic token roles rather than introducing a second visual language. The full contract, evaluation tool, and Chrome review skill live in [`design-system/`](design-system/) and [`.agents/skills/sloper-design-review/`](../../.agents/skills/sloper-design-review/).

## Views

| Route          | What it shows                                                        |
| -------------- | -------------------------------------------------------------------- |
| `/`            | Aggregate stats, pipeline funnel, failures, runs, and live activity   |
| `/sessions`    | Live pi transcripts grouped by issue or pull request                 |
| `/sessions/[id]` | Read-only transcript: prompts, thinking, tool calls, results        |
| `/issues`      | Searchable/filterable issue table, click through to detail           |
| `/issues/[id]` | Pipeline DAG, spec analysis, runs timeline, comments, event timeline |
| `/pulls`       | PR board grouped by GitHub state with agent review status             |
| `/runs`        | Filterable execution history with expandable agent output/thinking    |
| `/events`      | Hourly activity chart and filterable audit log stream                 |
| `/repos`       | Attach a repo link; the controller starts a docker instance for it     |
| `/instances`   | Manage multiple sloper instances + health probes                     |
| `/roadmap`     | Clearly labeled previews of planned loops, runners, and extensions    |

## Repositories

`/repos` is the one write path in the console: it talks to
[`sloper-controller`](../docs/repo-controller.md), which creates one container
per attached repo. **New repo** takes a link (`owner/repo` or a GitHub URL),
attaches it, and shows the container's status, port and logs. The repo's own
dashboard API is registered as an instance (without stealing the active one),
so **Open** switches the console to that repo's issues, runs and sessions. The
instance list defaults to `NEXT_PUBLIC_SLOPER_CONTROLLER_URL`
(`http://localhost:9090`) and can be changed per browser on the page.

## Sessions

The **Sessions** view tails the pi session files the workers write inside their
container (`~/.sloper/sessions`, served read-only by `sloper-web` through
`GET /api/sessions`, `/api/sessions/{id}` and `/api/sessions/{id}/events`).
There is no chat input: the console shows what a worker is doing, it never
sends anything back.

Because the file name carries sloper's deterministic session id
(`<timestamp>_sloper-issue-42-work.jsonl`), sessions are grouped by issue and by
the pull request that issue produced without any extra bookkeeping. The newest
entry decides the "current activity" line — an assistant message with a tool
call and no matching result means that tool is running right now.

pi writes an entry only when a step completes, so the transcript trails the
worker by the duration of its current step (a long `bash` call shows nothing
until it returns), and the file itself appears only after the first model
response. Transcripts disappear from the view when sloper deletes them: on
`/sloper approve` (spec session), on `/sloper abort`, and when the pull request
closes.

## Worktrees

The sidebar lists the worktrees the active instance is working in, split into
**Working now** (a worktree directory exists) and **Open work** (an issue has a
branch and a PR that is neither closed nor merged).

Sloper creates a worktree when a stage starts and deletes it when the stage
ends, so a directory on disk means an agent is in it at that moment. The
database only learns the branch name once the agent has finished, so the API
reads the instance's worktree directory and recovers the issue number from the
directory name (`sloper/issue-42-null-pointer`, `review-42-pr-118`) before
joining the database for the title and stage. Nothing is stored for this — the
list is rebuilt on every request.

A directory with no run behind it is flagged **orphaned**: the stage that
created it was killed before its cleanup. Sloper removes these on the next
start.

Set `SLOPER_WORKTREE_DIR` when the API server's home directory differs from
the one sloper ran under, as when the two live in separate containers.

## Running

```bash
# 1. Build the per-instance API server
make build-web          # -> dist/sloper-web

# 2. Run it next to each sloper instance (defaults: ~/.sloper/sloper.sqlite, :8080)
SLOPER_DB_PATH=/path/to/sloper.sqlite SLOPER_WEB_PORT=8080 ./dist/sloper-web

# Optional hardening: bind to localhost only by default; set a bearer token to
# require auth on every request (add the same token to the instance in the UI).
SLOPER_WEB_TOKEN=change-me SLOPER_WEB_ADDR=127.0.0.1 ./dist/sloper-web

# 3. Build + run the console
make build-dashboard    # npm install + next build
make run-dashboard      # next start on :3000
# or hot-reload during development
make dev-dashboard

# one shot
make dashboard
```

Open http://localhost:3000, add your instances on the **Instances** page, and the
console will start polling them.

For deployments where the API is not on the default `http://localhost:8080`, set
`NEXT_PUBLIC_SLOPER_DEFAULT_URL` before building the dashboard (for example,
`NEXT_PUBLIC_SLOPER_DEFAULT_URL=http://sloper-api:8080 make build-dashboard`). The
value seeds the first-run instance card; saved instances still take precedence. See
[`frontend/.env.example`](frontend/.env.example) for a copyable template.

## API

The Go server (`app/web/main.go`) exposes read-only JSON over the sloper SQLite DB.
By default it binds to `127.0.0.1:8080`. Set `SLOPER_WEB_TOKEN` to require a
`Authorization: Bearer <token>` header — the console stores per-instance tokens and
sends them automatically.

```
GET /api/health                instance health + repo name
GET /api/repo                  repo meta, db size, sanitized agent config
GET /api/summary               aggregate counts (issues/runs/pulls/events)
GET /api/issues                cached issues (?limit, ?stage, ?q search)
GET /api/issues/{n}            detail + spec + comments + runs + events + PR
GET /api/issues/{n}/comments   comments
GET /api/issues/{n}/runs       pipeline runs
GET /api/issues/{n}/events     event log for the issue
GET /api/pulls                 cached pull requests (?limit, ?offset, ?before_number)
GET /api/worktrees              worktree directories on disk + issues with unmerged work
GET /api/runs                  run history (?limit, ?offset, ?before_id)
GET /api/events                audit log (?limit, ?offset, ?before_id)
GET /api/metrics/activity      activity buckets for charts (?hours=24)
```

CORS is wide open (`*`) so the browser-based console can reach instances anywhere.
Pair that with `SLOPER_WEB_TOKEN` (or a reverse proxy with auth) whenever the server
is exposed beyond localhost.