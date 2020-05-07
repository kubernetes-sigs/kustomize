# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.13-stretch
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
RUN go build -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
CMD ["function"]
