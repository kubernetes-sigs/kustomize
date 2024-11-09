# Credit to Julien Guyomard (https://github.com/jguyomard)
# Credit to the Kubernetes Website team.
# This Dockerfile is based on:
# (https://github.com/kubernetes/website/blob/main/Dockerfile)

FROM docker.io/library/golang:1.22-alpine

RUN apk add --no-cache \
    curl \
    gcc \
    g++ \
    musl-dev \
    build-base \
    libc6-compat

RUN mkdir -p /go/src/sigs.k8s.io/kustomize/hack
WORKDIR /go/src/sigs.k8s.io/kustomize/hack
COPY hack/go.mod ./
COPY hack/go.sum ./
RUN go install --tags extended github.com/gohugoio/hugo

FROM docker.io/library/golang:1.22-alpine

RUN apk add --no-cache \
    runuser \
    git \
    openssh-client \
    rsync \
    npm

RUN mkdir -p /usr/local/node_packages
WORKDIR /usr/local/node_packages
COPY site/package.json ./
COPY site/package-lock.json ./
RUN npm install -D autoprefixer postcss-cli
ENV PATH="/usr/local/node_packages/node_modules/.bin:${PATH}"

RUN mkdir -p /var/hugo && \
    addgroup -Sg 1000 hugo && \
    adduser -Sg hugo -u 1000 -h /var/hugo hugo && \
    chown -R hugo: /var/hugo && \
    runuser -u hugo -- git config --global --add safe.directory /src

COPY --from=0 /go/bin/hugo /usr/local/bin/hugo

WORKDIR /src

USER hugo:hugo

EXPOSE 1313
