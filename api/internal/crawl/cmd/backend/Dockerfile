# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.11 AS build

ARG GO111MODULE=on

WORKDIR /go/src/sigs.k8s.io/kustomize/api/internal/crawl
COPY . /go/src/sigs.k8s.io/kustomize/api/internal/crawl

RUN go mod download
RUN CGO_ENABLED=0 go install sigs.k8s.io/kustomize/api/internal/crawl/cmd/backend/

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/backend /
ENTRYPOINT ["/backend"]
