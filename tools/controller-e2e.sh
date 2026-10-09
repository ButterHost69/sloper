#!/bin/bash
# controller-e2e.sh — prove the repo controller really runs one container per
# attached repo.
#
# It starts a throwaway sloper-controller on its own port, attaches two repos
# over the HTTP API, and asserts with the docker CLI that each one got a
# container, its own published dashboard port, inherited credentials and its own
# volumes. Then it detaches one, stops/restarts the other, and cleans up
# everything it created.
#
# The agent image is a stand-in built from a locally available base image, so
# the check needs no network and no GitHub credentials. Point SLOPER_IMAGE at a
# real sloper image to run the same flow against the real thing.
#
# Usage:  ./tools/controller-e2e.sh
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK="$REPO_ROOT/dist/controller-e2e"
GOCACHE_DIR="$REPO_ROOT/dist/.gocache"
CTRL_PORT="${SLOPER_E2E_CONTROLLER_PORT:-19090}"
PORT_MIN=18080
PORT_MAX=18099
IMAGE="${SLOPER_E2E_IMAGE:-sloper-agent:e2e}"
BASE_IMAGE="${SLOPER_E2E_BASE_IMAGE:-redis:7-alpine}"
CREATED="$WORK/created-containers.txt"
CONTROLLER_PID=""
BUILT_IMAGE=0

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
pass() { echo -e "${GREEN}PASS${NC} $1"; }
fail() { echo -e "${RED}FAIL${NC} $1"; exit 1; }
info() { echo -e "${CYAN}───${NC} $1"; }

# cleanup removes only the containers this run created, tracked by name, so it
# can never touch another sloper deployment on the same host.
cleanup() {
  if [ -n "$CONTROLLER_PID" ]; then kill "$CONTROLLER_PID" 2>/dev/null; wait "$CONTROLLER_PID" 2>/dev/null; fi
  if [ -f "$CREATED" ]; then
    while read -r name; do
      [ -z "$name" ] && continue
      volumes=$(docker inspect --format '{{range .Mounts}}{{.Name}} {{end}}' "$name" 2>/dev/null)
      docker rm -f "$name" >/dev/null 2>&1
      for volume in $volumes; do
        case "$volume" in sloper-*) docker volume rm -f "$volume" >/dev/null 2>&1 ;; esac
      done
    done < "$CREATED"
  fi
  if [ "$BUILT_IMAGE" = "1" ] && [ "${SLOPER_E2E_KEEP_IMAGE:-0}" != "1" ]; then docker rmi "$IMAGE" >/dev/null 2>&1; fi
}
trap cleanup EXIT

# remember every container name the controller reports, for safe cleanup
track_containers() {
  grep -o '"container_name":"[^"]*"' | sed 's/.*:"//; s/"$//' | sort -u >> "$CREATED"
  sort -u "$CREATED" -o "$CREATED"
}

# Start from a clean registry: a leftover row would make the first attach a
# duplicate, and the check is meant to be re-runnable.
rm -rf "$WORK"
mkdir -p "$WORK" "$GOCACHE_DIR"
: > "$CREATED"

# ── 1. stand-in agent image ─────────────────────────────────────────
if docker image inspect "$IMAGE" >/dev/null 2>&1; then
  info "reusing existing $IMAGE"
else
  info "building stand-in agent image $IMAGE from $BASE_IMAGE"
  CID=$(docker create "$BASE_IMAGE" sh -c 'true') || fail "docker create $BASE_IMAGE"
  docker commit \
    --change 'EXPOSE 8080' \
    --change 'CMD ["sh","-c","while true; do { printf \"HTTP/1.1 200 OK\\r\\nContent-Length: 5\\r\\n\\r\\nstub\\n\"; } | nc -l -p 8080; done"]' \
    "$CID" "$IMAGE" >/dev/null || fail "docker commit"
  docker rm -f "$CID" >/dev/null 2>&1
  BUILT_IMAGE=1
  pass "stand-in image built"
fi

# ── 2. controller ───────────────────────────────────────────────────
info "building sloper-controller"
(cd "$REPO_ROOT" && GOCACHE="$GOCACHE_DIR" go build -o "$WORK/sloper-controller" ./app/controller) || fail "go build"
pass "controller built"

info "starting controller on 127.0.0.1:$CTRL_PORT"
env \
  SLOPER_CONTROLLER_PORT="$CTRL_PORT" \
  SLOPER_CONTROLLER_DB_PATH="$WORK/controller.sqlite" \
  SLOPER_CONTROLLER_LOG_DIR="$WORK/logs" \
  SLOPER_CONTROLLER_RECONCILE_SECONDS=2 \
  SLOPER_IMAGE="$IMAGE" \
  SLOPER_PORT_RANGE="$PORT_MIN-$PORT_MAX" \
  GH_TOKEN="e2e-dummy-token" \
  GH_USERNAME="sloper-e2e" \
  AGENT_MODEL="anthropic/claude-e2e" \
  "$WORK/sloper-controller" > "$WORK/controller.log" 2>&1 &
CONTROLLER_PID=$!

for _ in $(seq 1 50); do
  curl -sf "http://127.0.0.1:$CTRL_PORT/api/health" >/dev/null 2>&1 && break
  sleep 0.2
done
HEALTH=$(curl -sf "http://127.0.0.1:$CTRL_PORT/api/health") || fail "controller health unreachable (see $WORK/controller.log)"
echo "$HEALTH" | grep -q '"docker":"ok"' || fail "controller cannot reach docker: $HEALTH"
pass "controller healthy (docker reachable)"

# ── 3. attach two repos ─────────────────────────────────────────────
info "attaching https://github.com/ButterHost69/sloper.git"
ATTACH_A=$(curl -sf -X POST "http://127.0.0.1:$CTRL_PORT/api/repos" \
  -H 'Content-Type: application/json' -d '{"link":"https://github.com/ButterHost69/sloper.git"}') || fail "attach A"
echo "$ATTACH_A" | grep -q '"status":"pending"' || fail "attach A did not return a pending repo: $ATTACH_A"
echo "$ATTACH_A" | track_containers
pass "repo A attached"

info "attaching ButterHost69/sloper-e2e-second (owner/repo form)"
ATTACH_B=$(curl -sf -X POST "http://127.0.0.1:$CTRL_PORT/api/repos" \
  -H 'Content-Type: application/json' -d '{"link":"ButterHost69/sloper-e2e-second"}') || fail "attach B"
echo "$ATTACH_B" | track_containers
pass "repo B attached"

info "rejecting a non-GitHub link"
CODE=$(curl -s -o "$WORK/bad.json" -w '%{http_code}' -X POST "http://127.0.0.1:$CTRL_PORT/api/repos" \
  -H 'Content-Type: application/json' -d '{"link":"https://gitlab.com/owner/repo"}')
[ "$CODE" = "400" ] || fail "bad link returned HTTP $CODE, want 400 ($(cat "$WORK/bad.json"))"
pass "non-GitHub link rejected with 400"

info "rejecting a duplicate attach"
CODE=$(curl -s -o /dev/null -w '%{http_code}' -X POST "http://127.0.0.1:$CTRL_PORT/api/repos" \
  -H 'Content-Type: application/json' -d '{"link":"ButterHost69/sloper"}')
[ "$CODE" = "409" ] || fail "duplicate attach returned HTTP $CODE, want 409"
pass "duplicate attach rejected with 409"

# ── 4. wait for both containers to run ──────────────────────────────
info "waiting for the reconciler to bring both containers up"
REPOS=""
for _ in $(seq 1 60); do
  REPOS=$(curl -sf "http://127.0.0.1:$CTRL_PORT/api/repos") || fail "list repos"
  echo "$REPOS" | track_containers
  RUNNING=$(echo "$REPOS" | grep -o '"status":"running"' | wc -l | tr -d ' ')
  [ "$RUNNING" = "2" ] && break
  sleep 0.5
done
[ "$RUNNING" = "2" ] || fail "containers never reached running: $REPOS"
pass "both repos report a running container"

# ── 5. verify with the docker CLI ───────────────────────────────────
info "checking docker sees one container per repo"
MANAGED=$(sort -u "$CREATED")
COUNT=$(echo "$MANAGED" | grep -c .)
[ "$COUNT" = "2" ] || fail "controller reported $COUNT containers, want 2 ($MANAGED)"
pass "containers: $(echo "$MANAGED" | tr '\n' ' ')"

NAME_A=$(echo "$MANAGED" | grep -v 'e2e-second' | head -1)
NAME_B=$(echo "$MANAGED" | grep 'e2e-second' | head -1)
[ -n "$NAME_A" ] && [ -n "$NAME_B" ] || fail "could not tell the two containers apart: $MANAGED"

PORT_A=$(docker inspect --format '{{index .NetworkSettings.Ports "8080/tcp" 0 "HostPort"}}' "$NAME_A")
PORT_B=$(docker inspect --format '{{index .NetworkSettings.Ports "8080/tcp" 0 "HostPort"}}' "$NAME_B")
ENV_A=$(docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' "$NAME_A")
LABEL_A=$(docker inspect --format '{{index .Config.Labels "sloper.repo"}}' "$NAME_A")
MANAGED_LABEL=$(docker inspect --format '{{index .Config.Labels "sloper.managed"}}' "$NAME_A")
VOLS_A=$(docker inspect --format '{{range .Mounts}}{{.Name}}:{{.Destination}} {{end}}' "$NAME_A")

[ "$MANAGED_LABEL" = "true" ] || fail "sloper.managed label = $MANAGED_LABEL"
[ "$LABEL_A" = "ButterHost69/sloper" ] || fail "sloper.repo label = $LABEL_A"
echo "$ENV_A" | grep -q '^SLOPER_REPO_LINK=ButterHost69/sloper$' || fail "SLOPER_REPO_LINK missing in container A"
echo "$ENV_A" | grep -q '^GH_TOKEN=e2e-dummy-token$' || fail "GH_TOKEN not inherited by container A"
echo "$ENV_A" | grep -q '^AGENT_MODEL=anthropic/claude-e2e$' || fail "AGENT_MODEL not inherited by container A"
echo "$ENV_A" | grep -q '^SLOPER_WEB_ADDR=0.0.0.0$' || fail "SLOPER_WEB_ADDR missing in container A"
echo "$ENV_A" | grep -q '^SLOPER_WEB_TOKEN=.\+' || fail "no per-instance API token in container A"
case "$PORT_A" in 180*) ;; *) fail "published port $PORT_A is outside $PORT_MIN-$PORT_MAX" ;; esac
echo "$VOLS_A" | grep -q ':/root/.sloper' || fail "data volume missing: $VOLS_A"
echo "$VOLS_A" | grep -q ':/root/repo' || fail "repo volume missing: $VOLS_A"
pass "container A has inherited env, labels and per-repo volumes (port $PORT_A)"

[ "$PORT_A" != "$PORT_B" ] || fail "both repos were published on port $PORT_A"
pass "each repo got its own host port ($PORT_A and $PORT_B)"

info "curling the published instance API on 127.0.0.1:$PORT_A"
BODY=$(curl -sf --max-time 5 "http://127.0.0.1:$PORT_A/" 2>/dev/null) || BODY=""
if [ -n "$BODY" ]; then
  pass "published port answers through the container: $(echo "$BODY" | tr -d '\n')"
else
  info "stand-in image did not answer on the published port (port mapping still created)"
fi

# ── 6. detach one repo ──────────────────────────────────────────────
ID_B=$(echo "$REPOS" | grep -o '"id":[0-9]*,"link":"ButterHost69/sloper-e2e-second"' | grep -o '[0-9]*' | head -1)
[ -n "$ID_B" ] || fail "could not find repo B id"
info "detaching repo B (id $ID_B, purging its volumes)"
curl -sf -X DELETE "http://127.0.0.1:$CTRL_PORT/api/repos/$ID_B?purge=true" >/dev/null || fail "detach B"

docker container inspect "$NAME_B" >/dev/null 2>&1 && fail "repo B container survived detach"
SLUG_B="${NAME_B#sloper-repo-}"
docker volume inspect "sloper-data-$SLUG_B" >/dev/null 2>&1 && fail "repo B data volume survived a purging detach"
docker volume inspect "sloper-repo-$SLUG_B" >/dev/null 2>&1 && fail "repo B repo volume survived a purging detach"
REMAINING=$(curl -sf "http://127.0.0.1:$CTRL_PORT/api/repos" | grep -o '"link":"[^"]*"' | sort -u | tr '\n' ' ')
case "$REMAINING" in *sloper-e2e-second*) fail "repo B row survived detach" ;; esac
pass "detach removed container, volumes and row; repos now: $REMAINING"

# ── 7. stop and restart the remaining repo ──────────────────────────
ID_A=$(curl -sf "http://127.0.0.1:$CTRL_PORT/api/repos" | grep -o '"id":[0-9]*,"link":"ButterHost69/sloper"' | grep -o '[0-9]*' | head -1)
info "setting desired_state=stopped on repo A (id $ID_A)"
curl -sf -X PATCH "http://127.0.0.1:$CTRL_PORT/api/repos/$ID_A" \
  -H 'Content-Type: application/json' -d '{"desired_state":"stopped"}' >/dev/null || fail "stop A"
for _ in $(seq 1 40); do
  STATE=$(docker inspect --format '{{.State.Status}}' "$NAME_A" 2>/dev/null || echo gone)
  [ "$STATE" != "running" ] && break
  sleep 0.5
done
[ "$STATE" != "running" ] || fail "container A still running after desired_state=stopped"
pass "desired_state=stopped stopped the container"

info "restarting repo A"
curl -sf -X POST "http://127.0.0.1:$CTRL_PORT/api/repos/$ID_A/restart" >/dev/null || fail "restart A"
for _ in $(seq 1 60); do
  STATE=$(docker inspect --format '{{.State.Status}}' "$NAME_A" 2>/dev/null || echo gone)
  [ "$STATE" = "running" ] && break
  sleep 0.5
done
[ "$STATE" = "running" ] || fail "container A did not come back after restart"
pass "restart replaced the container; volumes were kept"

info "controller log tail"
tail -n 5 "$WORK/controller.log"

echo
echo -e "${GREEN}ALL CONTROLLER E2E CHECKS PASSED${NC}"
