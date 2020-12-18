# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.154.0/containers/ubuntu/.devcontainer/base.Dockerfile

# [Choice] Ubuntu version: bionic, focal
ARG VARIANT
FROM mcr.microsoft.com/vscode/devcontainers/base:0-${VARIANT}

# Arguments used during build
ARG GO_VERSION

# Install base OS packages
RUN set -ex \
    && apt-get update \
    && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        build-essential \
        gnupg2

# Install Go
RUN set -ex \
    && cd /usr/local \
    && wget -c https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go

RUN set -ex \
    && go get -v golang.org/x/tools/gopls

# Install Docker
RUN set -ex \
    && curl -fsSL https://get.docker.com -o get-docker.sh \
    && sh get-docker.sh \
    && usermod -aG docker vscode
