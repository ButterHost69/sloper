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

echo "== Starting Sloper =="
cp /usr/local/bin/sloper ./sloper
./sloper
