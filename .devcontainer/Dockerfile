# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.137.0/containers/go/.devcontainer/base.Dockerfile
ARG VARIANT="1"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Optional] Uncomment this section to install additional OS packages.
RUN set -ex \
    && apt-get update \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        gnupg2 \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg-agent \
        software-properties-common

# Install Docker
RUN set -ex \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add - \
    && apt-key fingerprint 0EBFCD88 \
    && add-apt-repository \
        "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
        focal \
        stable" \
    && apt-get update \
    && apt-get  -y install --no-install-recommends \
        build-essential
        docker-ce \
        docker-ce-cli \
        containerd.io \
    && groupadd docker \
    && usermod -aG docker vscode \
    && systemctl enable docker
