# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.21-bullseye
ENV CGO_ENABLED=0
RUN mkdir -p /go/src/sigs.k8s.io/kustomize/functions/examples/template-heredoc-cockroachdb/image/tools/
WORKDIR /go/src/sigs.k8s.io/kustomize/functions/examples/template-heredoc-cockroachdb/image/tools/
COPY tools/go.mod ./
COPY tools/go.sum ./
RUN BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    && KUSTOMIZE_VERSION="$(grep sigs.k8s.io/kustomize/kustomize/v5 go.mod | cut -d' ' -f3)" \
    && go install -ldflags \
    "-X sigs.k8s.io/kustomize/api/provenance.buildDate=${BUILD_DATE} \
    -X sigs.k8s.io/kustomize/api/provenance.version=${KUSTOMIZE_VERSION}" \
    sigs.k8s.io/kustomize/kustomize/v5 \
    && echo "kustomize version: $(kustomize version)"

FROM alpine:latest
RUN apk add --no-cache bash
COPY cockroachdb-template.sh /usr/local/bin/config-function
CMD ["config-function"]
