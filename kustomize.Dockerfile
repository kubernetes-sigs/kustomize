# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# build
FROM public.ecr.aws/docker/library/golang:1.22.7-bullseye AS builder
ARG VERSION
ARG DATE
RUN mkdir /build
ADD . /build/
WORKDIR /build/kustomize
RUN CGO_ENABLED=0 GO111MODULE=on go build \
    -ldflags="-s -X sigs.k8s.io/kustomize/api/provenance.version=${VERSION} \
    -X sigs.k8s.io/kustomize/api/provenance.buildDate=${DATE}"

# only copy binary
FROM public.ecr.aws/docker/library/alpine
# install dependencies
RUN apk add --no-cache git openssh
COPY --from=builder /build/kustomize/kustomize /app/
WORKDIR /app
ENV PATH="$PATH:/app"
ENTRYPOINT ["/app/kustomize"]
