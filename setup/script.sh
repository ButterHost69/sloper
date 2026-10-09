#!/bin/bash
# Entrypoint for a repo container. sloper-controller starts one of these per
# attached repo and assigns the repo through SLOPER_REPO_LINK; no .env file
# selects a repo any more, so a container without that variable is a mistake.
set -e

# Source nvm so node/npm/pi are available
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

echo 'export PATH="$HOME/.npm-global/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc

: "${SLOPER_REPO_LINK:?SLOPER_REPO_LINK is required - repos are attached through sloper-controller}"

cd ~/repo
git config --global user.name "$GH_USERNAME"
git config --global user.email "$GH_EMAIL"

if [ -d ".git" ]; then
	echo "== Repo exists, pulling latest changes =="
	git fetch --all
	git reset --hard origin/HEAD 2>/dev/null || true
	git clean -fd
else
	gh repo clone "$SLOPER_REPO_LINK" .
fi

gh auth setup-git

# Start the read-only dashboard API in the background so the console can
# observe this instance (SLOPER_DB_PATH defaults to ~/.sloper/sloper.sqlite).
# The container must bind 0.0.0.0 so the host-published port is reachable.
echo "== Starting Sloper Web API =="
SLOPER_DB_PATH="${SLOPER_DB_PATH:-$HOME/.sloper/sloper.sqlite}" \
SLOPER_SESSION_DIR="${SLOPER_SESSION_DIR:-$HOME/.sloper/sessions}" \
SLOPER_WEB_PORT="${SLOPER_WEB_PORT:-8080}" \
SLOPER_WEB_ADDR="${SLOPER_WEB_ADDR:-0.0.0.0}" \
SLOPER_REPO="${SLOPER_REPO:-$(basename "$SLOPER_REPO_LINK" .git)}" \
nohup sloper-web >/tmp/sloper-web.log 2>&1 &

echo "== Starting Sloper =="
cp /usr/local/bin/sloper ./sloper
./sloper
