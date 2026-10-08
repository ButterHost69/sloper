#!/bin/bash
set -e

# Source nvm so node/npm/pi are available
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

echo 'export PATH="$HOME/.npm-global/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc

cd ~/repo
git config --global user.name "$GH_USERNAME"
git config --global user.email "$GH_EMAIL"

if [ -d ".git" ]; then
	echo "== Repo exists, pulling latest changes =="
	git fetch --all
	git reset --hard origin/HEAD 2>/dev/null || true
	git clean -fd
else
	gh repo clone $GH_REPO_LINK .
fi

gh auth setup-git

# Start the read-only dashboard API in the background so the console can
# observe this instance (SLOPER_DB_PATH defaults to ~/.sloper/sloper.sqlite).
# The container must bind 0.0.0.0 so the host-published port is reachable.
echo "== Starting Sloper Web API =="
SLOPER_DB_PATH="${SLOPER_DB_PATH:-$HOME/.sloper/sloper.sqlite}" \
SLOPER_WEB_PORT="${SLOPER_WEB_PORT:-8080}" \
SLOPER_WEB_ADDR="${SLOPER_WEB_ADDR:-0.0.0.0}" \
SLOPER_REPO="${SLOPER_REPO:-$(basename "$GH_REPO_LINK" .git)}" \
nohup sloper-web >/tmp/sloper-web.log 2>&1 &

echo "== Starting Sloper =="
cp /usr/local/bin/sloper ./sloper
./sloper
