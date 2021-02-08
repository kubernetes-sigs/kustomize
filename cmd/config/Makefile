# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test build tidy

GOBIN := $(shell go env GOPATH)/bin

$(GOBIN)/addlicense:
	go get github.com/google/addlicense

$(GOBIN)/golangci-lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.19.1

$(GOBIN)/mdtogo:
	go get sigs.k8s.io/kustomize/cmd/mdtogo

build:
	go build -v -o $(GOBIN)/kubectl-krm ./kubectl-krm

all: build license fix vet fmt test lint tidy

generate: $(GOBIN)/mdtogo
	GOBIN=$(GOBIN) go generate ./...

lint: $(GOBIN)/golangci-lint
	$(GOBIN)/golangci-lint run ./...

license: $(GOBIN)/addlicense
	$(GOBIN)/addlicense \
	  -y 2021 \
	  -c "The Kubernetes Authors." \
	  -f LICENSE_TEMPLATE .

test:
	go test -v -timeout 45m -cover ./...

fix:
	go fix ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...
