# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN go build -v -o /usr/local/bin/config-function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["config-function"]