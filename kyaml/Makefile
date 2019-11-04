# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate fix vet fmt test build tidy

all: generate fix vet fmt test build tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	go generate ./...

tidy:
	go mod tidy

lint:
	(which $(go env GOPATH)/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	$(go env GOPATH)/bin/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...

