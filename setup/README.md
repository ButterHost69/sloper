# Sloper Docker Setup

Sets up sloper's backend in Docker. Copy the env template and fill it in:

```bash
cp setup/.env.example setup/.env     # GH_TOKEN, AGENT_MODEL, SLOPER_CONTROLLER_TOKEN
make docker                          # builds the image and starts the controller
# controller API is now reachable at http://localhost:9090/api/health
```

The compose stack runs **`sloper-controller`**. Attach repos through the
controller — the console's **New repo** button, or `POST /api/repos` — and it
starts one container per attached repo from the same image, each with its own
host port and its own two volumes. See
[docs/repo-controller.md](../docs/repo-controller.md) for the API, the lifecycle
rules and every environment variable.

The controller needs the docker socket (`compose.yml` mounts
`/var/run/docker.sock`) because starting those containers is its job, and it
refuses to serve an unauthenticated API off loopback — hence the required
`SLOPER_CONTROLLER_TOKEN`. Repo instances are published on
`SLOPER_PORT_RANGE` (default `8080-8180`); keep `SLOPER_PUBLISH_ADDR` and
`SLOPER_PUBLIC_HOST` in step with where the console runs.

To run the controller on the host instead of in Docker, use
`make run-controller` — same binary, same API, no socket mount needed.

> After changing the Go source, rebuild with `make docker`: the binaries are
> baked into the image via the `Dockerfile`.

## What a repo container does

`setup/script.sh` is the image's default entrypoint and the entrypoint of every
repo container. It sources nvm, sets the git identity from `GH_USERNAME`/
`GH_EMAIL`, clones the repo the controller assigned through `SLOPER_REPO_LINK`
(or pulls it when the volume already holds a clone), then starts `sloper-web` in
the background on `0.0.0.0:8080` and `sloper` in the foreground. Two named
volumes per repo keep the clone (`/root/repo`) and the state — database,
sessions, worktrees (`/root/.sloper`) — across restarts.

A repo container is created by the controller, which assigns `SLOPER_REPO_LINK`;
starting the image by hand without that variable fails immediately.

## Agent session transcripts

The console's **Sessions** view reads the pi session files the workers write to
`$HOME/.sloper/sessions`, which lives inside the repo's `/root/.sloper` volume.
Each repo container serves its own transcripts through its own `sloper-web`, so
no extra mount is needed.

Each stage gets a deterministic pi session id (`sloper-issue-42-work`), so the
file name alone tells the console which issue, stage, and pull request a
transcript belongs to.

## Browser tools for pi workers

The image ships `pi-mcp-adapter` and a chrome-devtools MCP server config
(`/root/.pi/agent/mcp.json`) so pi workers can drive a headless Chrome inside
the container (screenshots, console, network, performance). No extra ports
needed; the browser can reach that repo's sloper-web at http://localhost:8080
from inside its own container.
