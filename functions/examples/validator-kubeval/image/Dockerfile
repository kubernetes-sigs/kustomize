# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

FROM alpine:latest as schemas
RUN apk --no-cache add git
RUN git clone --depth 1 https://github.com/instrumenta/kubernetes-json-schema.git

FROM public.ecr.aws/docker/library/golang:1.22.7-bullseye as function
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
RUN go build -v -o /usr/local/bin/config-function ./

FROM alpine:latest
COPY --from=schemas /kubernetes-json-schema/v1.16.0-standalone-strict /schemas/v1.16.0-standalone-strict
COPY --from=function /usr/local/bin/config-function /usr/local/bin/config-function
CMD ["config-function"]
