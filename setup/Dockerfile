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

# Download and Install agent (pi)
# RUN curl -fsSL https://opencode.ai/install | bash

RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.5/install.sh | bash
ENV NVM_DIR=/root/.nvm
RUN \. "$HOME/.nvm/nvm.sh"
RUN . "$NVM_DIR/nvm.sh" && nvm install 24
RUN . "$NVM_DIR/nvm.sh" && node -v # Should print "v24.18.0".
RUN . "$NVM_DIR/nvm.sh" && npm -v # Should print "11.16.0".
RUN . "$NVM_DIR/nvm.sh" && npm install -g --ignore-scripts @earendil-works/pi-coding-agent \
    && ln -sf "$(npm root -g)/@earendil-works/pi-coding-agent/dist/cli.js" /usr/local/bin/pi \
    && chmod +x /usr/local/bin/pi

COPY ./script.sh ./setup.sh
COPY ./sloper /usr/local/bin/sloper

RUN mkdir -p ~/repo
RUN chmod +x /setup.sh
RUN chmod +x /usr/local/bin/sloper

ENTRYPOINT ["bash", "/setup.sh"]
CMD []