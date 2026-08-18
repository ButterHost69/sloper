FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl wget git vim nano \
    build-essential pkg-config \
    ca-certificates gnupg \
    unzip jq tree htop \
    && rm -rf /var/lib/apt/lists/*


RUN (type -p wget >/dev/null || ( apt update &&  apt install wget -y)) \
	&&  mkdir -p -m 755 /etc/apt/keyrings \
	&& out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
	&& cat $out |  tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
	&&  chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
	&&  mkdir -p -m 755 /etc/apt/sources.list.d \
	&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" |  tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
	&&  apt update \
	&&  apt install gh -y

# Install Google Chrome for the chrome-devtools MCP server used by pi workers.
RUN wget -q -O /tmp/google-chrome.deb https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
	&& apt-get update \
	&& apt-get install -y --no-install-recommends /tmp/google-chrome.deb \
	&& rm /tmp/google-chrome.deb \
	&& rm -rf /var/lib/apt/lists/*

# Download and Install agent (pi)
# RUN curl -fsSL https://opencode.ai/install | bash

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.5/install.sh | bash
ENV NVM_DIR=/root/.nvm
RUN \. "$HOME/.nvm/nvm.sh"
RUN . "$NVM_DIR/nvm.sh" && nvm install 24
RUN . "$NVM_DIR/nvm.sh" && node -v # Should print "v24.18.0".
RUN . "$NVM_DIR/nvm.sh" && npm -v # Should print "11.16.0".
RUN . "$NVM_DIR/nvm.sh" && npm install -g --ignore-scripts @earendil-works/pi-coding-agent@0.84.2 \
    && ln -sf "$(npm root -g)/@earendil-works/pi-coding-agent/dist/cli.js" /usr/local/bin/pi \
    && chmod +x /usr/local/bin/pi

# Install the pi MCP adapter so pi workers can use MCP servers (chrome-devtools).
# Primary: `pi install` writes settings.json and fetches the package.
# Fallback: if `pi install` fails (e.g. no TTY in the build), install the
# package into pi's user package dir and write settings.json manually.
RUN set -e; . "$NVM_DIR/nvm.sh"; \
	pi install npm:pi-mcp-adapter@2.26.1 \
		&& echo "== pi-mcp-adapter: installed via 'pi install' ==" \
		|| echo "== pi-mcp-adapter: 'pi install' failed (exit $?), checking fallback =="; \
	if [ -d /root/.pi/agent/npm/node_modules/pi-mcp-adapter ] \
		&& grep -q pi-mcp-adapter /root/.pi/agent/settings.json 2>/dev/null; then \
		echo "== pi-mcp-adapter: package and settings present =="; \
	else \
		echo "== pi-mcp-adapter: using manual fallback install =="; \
		mkdir -p /root/.pi/agent/npm; \
		cd /root/.pi/agent/npm; \
		npm init -y >/dev/null 2>&1; \
		npm install pi-mcp-adapter@2.26.1; \
		printf '%s\n' '{"packages": ["pi-mcp-adapter@2.26.1"]}' > /root/.pi/agent/settings.json; \
	fi

COPY ./script.sh ./setup.sh
COPY ./sloper /usr/local/bin/sloper
COPY ./skills/ /root/.pi/agent/skills/
COPY ./pi-mcp.json /root/.pi/agent/mcp.json

RUN mkdir -p ~/repo
RUN chmod +x /setup.sh
RUN chmod +x /usr/local/bin/sloper

ENTRYPOINT ["bash", "/setup.sh"]
CMD []