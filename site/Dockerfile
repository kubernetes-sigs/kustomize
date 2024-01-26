# Credit to Julien Guyomard (https://github.com/jguyomard)
# Credit to the Kubernetes Website team.
# This Dockerfile is based on:
# (https://github.com/kubernetes/website/blob/main/Dockerfile)

FROM docker.io/library/golang:1.20-alpine

RUN apk add --no-cache \
    curl \
    gcc \
    g++ \
    musl-dev \
    build-base \
    libc6-compat

ARG HUGO_VERSION

RUN mkdir $HOME/src && \
    cd $HOME/src && \
    curl -L https://github.com/gohugoio/hugo/archive/refs/tags/v${HUGO_VERSION}.tar.gz | tar -xz && \
    cd "hugo-${HUGO_VERSION}" && \
    go install --tags extended

FROM docker.io/library/golang:1.20-alpine

RUN apk add --no-cache \
    runuser \
    git \
    openssh-client \
    rsync \
    npm && \
    npm install -D autoprefixer postcss-cli

RUN mkdir -p /var/hugo && \
    addgroup -Sg 1000 hugo && \
    adduser -Sg hugo -u 1000 -h /var/hugo hugo && \
    chown -R hugo: /var/hugo && \
    runuser -u hugo -- git config --global --add safe.directory /src

COPY --from=0 /go/bin/hugo /usr/local/bin/hugo

WORKDIR /src

USER hugo:hugo

EXPOSE 1313
