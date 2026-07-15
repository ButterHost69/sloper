#!/bin/bash
# validate.sh — smoke-test each layer of Sloper
# Usage:  ./validate.sh [repo-path]
set -euo pipefail

REPO="${1:-$(pwd)}"
RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
pass() { echo -e "${GREEN}PASS${NC} $1"; }
fail() { echo -e "${RED}FAIL${NC} $1"; exit 1; }
warn() { echo -e "${RED}WARN${NC} $1"; }
info() { echo -e "${CYAN}───${NC} $1"; }

# ── 0. prerequisites ────────────────────────────────────────────────
info "Checking prerequisites"
command -v go   >/dev/null || fail "go not found"
command -v gh   >/dev/null || fail "gh not found"
command -v pi   >/dev/null || fail "pi not found"
command -v git  >/dev/null || fail "git not found"
pass "go $(go version | awk '{print $3}'), gh $(gh --version | head -1 | awk '{print $3}'), pi $(pi --version), git $(git --version | awk '{print $3}')"

# ── 1. build ────────────────────────────────────────────────────────
info "Building sloper"
cd "$(dirname "$0")"
go build -o /tmp/sloper-test ./app/sloper 2>&1 || fail "build failed"
pass "sloper binary built → /tmp/sloper-test"

# ── 2. compile check ────────────────────────────────────────────────
info "Checking all packages compile"
go vet ./internal/... 2>&1 && pass "all packages vet clean" || fail "vet errors"

# ── 3. gh auth ──────────────────────────────────────────────────────
info "Checking GitHub CLI auth"
if gh auth status 2>&1 | grep -q "Logged in"; then
    pass "gh authenticated"
else
    warn "gh not authenticated — run: gh auth login"
fi

# ── 4. detect repo ──────────────────────────────────────────────────
info "Detecting GitHub repo from $REPO"
REPO_NAME=$(cd "$REPO" && git config --get remote.origin.url 2>/dev/null || true)
if [ -n "$REPO_NAME" ]; then
    pass "git remote: $REPO_NAME"
else
    warn "no git remote found in $REPO — sloper needs a GitHub repo"
fi

# ── 5. API key check ────────────────────────────────────────────────
info "Checking AI provider credentials"
if [ -n "${ANTHROPIC_API_KEY:-}" ]; then
    pass "ANTHROPIC_API_KEY is set"
elif [ -n "${OPENAI_API_KEY:-}" ]; then
    pass "OPENAI_API_KEY is set"
else
    warn "No API key found (ANTHROPIC_API_KEY / OPENAI_API_KEY)"
    echo "  Sloper needs an API key to drive pi. Options:"
    echo "  1. export ANTHROPIC_API_KEY=sk-ant-..."
    echo "  2. export OPENAI_API_KEY=sk-..."
    echo "  3. pi /login (interactive OAuth for supported providers)"
    echo ""
    echo "  If you use pi interactively with a subscription (opencode-go, copilot, etc.),"
    echo "  set AGENT_MODEL to a model from that provider:"
    echo "    export AGENT_MODEL='opencode-go/kimi-k2.7-code'"
fi

# ── 6. RPC mode check ───────────────────────────────────────────────
info "Testing pi --mode rpc (quick smoke test — 5s timeout)"
RPC_LINES=$(echo '{"type":"prompt","message":"Say OK"}' | timeout 5 pi --mode rpc --no-session --no-extensions --no-skills --no-context-files --thinking off 2>/dev/null | wc -l)
if [ "$RPC_LINES" -gt 3 ]; then
    pass "pi RPC mode produces output ($RPC_LINES lines)"
else
    warn "pi RPC produced only $RPC_LINES lines — may need an API key"
fi

# ── 7. run instructions ─────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  READY TO RUN"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  cd $REPO"
echo "  export GH_TOKEN=\$(gh auth token)"
echo "  export AGENT_MODEL='anthropic/claude-sonnet-4-20250514'   # or your model"
echo "  /tmp/sloper-test"
echo ""
echo "  The scheduler will:"
echo "    1. Detect the repo from git remote"
echo "    2. List open issues"
echo "    3. For each untriaged issue: spawn pi, run SPEC stage"
echo "    4. Log the spec summary and files to change"
echo ""
echo "  Output goes to stderr (log.Printf).  Pipe through tee to save:"
echo "    /tmp/sloper-test 2>&1 | tee sloper-run.log"
echo ""
