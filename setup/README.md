# Sloper Docker Setup

Setups Sloper in a Docker Environment.
Add GH Token & Repo Name in the .env file (rename the .env.example -> .env)

This compose file runs **one repo**. To run several, attach them from the
console instead: [`sloper-controller`](../docs/repo-controller.md) starts one
container per attached repo from the same image, with its own port and volumes.

```bash
make build-agent-image   # this image, tagged sloper-agent:latest
make run-controller      # then "New repo" in the console
```

Each container also runs the read-only **dashboard API** (`sloper-web`) on port
`8080` so the [console](../dashboard/README.md) can observe the instance:

```bash
make docker              # builds sloper + sloper-web image and starts it
# dashboard API is now reachable at http://localhost:8080/api/health
```

To change the host port, set `SLOPER_WEB_PORT` in `.env` (e.g. `8081` for a
second instance) — it is forwarded by `compose.yml`. The web server binds
`0.0.0.0` inside the container (set via `SLOPER_WEB_ADDR`) so the published port
is reachable from the host; you can override it with `SLOPER_WEB_ADDR` in `.env`.
To require a bearer token, set `SLOPER_WEB_TOKEN` in `.env` and add the same
token to the instance in the console UI.

> After changing the Go source (e.g. `app/web/main.go`), rebuild the image with
> `make docker` — the binaries are baked into the image via the `Dockerfile`.

## Agent session transcripts

The console's **Sessions** view reads the pi session files the workers write to
`$HOME/.sloper/sessions`, which lives inside the `sloper_data` volume that is
already mounted at `/root/.sloper`. `sloper-web` therefore needs no extra mount;
`SLOPER_SESSION_DIR` (default `/root/.sloper/sessions`, set in `compose.yml`)
only has to change when the API server and the workers do not share a home
directory.

Each stage gets a deterministic pi session id (`sloper-issue-42-work`), so the
file name alone tells the console which issue, stage, and pull request a
transcript belongs to.

## Browser tools for pi workers

The image ships `pi-mcp-adapter` and a chrome-devtools MCP server config
(`/root/.pi/agent/mcp.json`) so pi workers can drive a headless Chrome inside
the container (screenshots, console, network, performance). No extra ports
needed; the browser can reach sloper-web at http://localhost:8080.
