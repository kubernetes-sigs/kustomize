# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.13-stretch
ENV CGO_ENABLED=0
RUN go get -v sigs.k8s.io/kustomize/kustomize

FROM alpine:latest
RUN apk add --no-cache bash
COPY --from=0 /go/bin/kustomize /usr/local/bin
COPY cockroachdb-template.sh /usr/local/bin/config-function
CMD ["config-function"]