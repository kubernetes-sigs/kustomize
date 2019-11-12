# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test lint tidy

GOPATH := $(shell go env GOPATH)

all: generate license fix vet fmt test lint tidy

fix:
	go fix ./...

fmt:
	go fmt ./...

generate:
	go generate ./...

license:
	(which $(GOPATH)/bin/addlicense || go get github.com/google/addlicense)
	$(GOPATH)/bin/addlicense  -y 2019 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

tidy:
	go mod tidy

lint:
	(which $(GOPATH)/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1)
	$(GOPATH)/bin/golangci-lint run ./...

test:
	go test -cover ./...

vet:
	go vet ./...
