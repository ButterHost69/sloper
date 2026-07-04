#!/bin/bash
set -e

cd ~/repo
git config --global user.name "$GH_USERNAME"
git config --global user.email "$GH_EMAIL"

gh repo clone $GH_REPO_LINK .

echo "== Starting Sloper =="
mv /usr/local/bin/sloper ./sloper
./sloper
# /usr/local/bin/sloper
