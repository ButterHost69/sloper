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


COPY ./script.sh ./setup.sh
COPY ./sloper /usr/local/bin/sloper

RUN mkdir -p ~/repo
RUN chmod +x /setup.sh
RUN chmod +x /usr/local/bin/sloper

ENTRYPOINT ["bash", "/setup.sh"]
CMD []