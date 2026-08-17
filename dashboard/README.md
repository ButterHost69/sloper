# Sloper Console

A dark-mode observability dashboard for **one or more running sloper instances**,
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

## Views

| Route          | What it shows                                                        |
| -------------- | -------------------------------------------------------------------- |
| `/`            | Aggregate stats, pipeline funnel, activity chart, live event feed    |
| `/issues`      | Searchable/filterable issue table, click through to detail           |
| `/issues/[id]` | Pipeline DAG, spec analysis, runs timeline, comments, event timeline |
| `/pulls`       | PR board grouped by state with review status                         |
| `/runs`        | Pipeline execution history with expandable agent output/thinking     |
| `/events`      | Filterable audit log stream                                          |
| `/instances`   | Manage multiple sloper instances + health probes                     |

## Running

```bash
# 1. Build the per-instance API server
make build-web          # -> dist/sloper-web

# 2. Run it next to each sloper instance (defaults: ~/.sloper/sloper.sqlite, :8080)
SLOPER_DB_PATH=/path/to/sloper.sqlite SLOPER_WEB_PORT=8080 ./dist/sloper-web

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

## API

The Go server (`app/web/main.go`) exposes read-only JSON over the sloper SQLite DB:

```
GET /api/health                instance health + repo name
GET /api/repo                  repo meta, db size, sanitized agent config
GET /api/summary               aggregate counts (issues/runs/pulls/events)
GET /api/issues                cached issues (limit/offset)
GET /api/issues/{n}            detail + spec + comments + runs + events + PR
GET /api/issues/{n}/comments   comments
GET /api/issues/{n}/runs       pipeline runs
GET /api/issues/{n}/events     event log for the issue
GET /api/pulls                 cached pull requests
GET /api/runs                  all runs
GET /api/events                audit log
GET /api/metrics/activity      activity buckets for charts (?hours=24)
```

CORS is wide open (`*`) so the browser-based console can reach instances anywhere.