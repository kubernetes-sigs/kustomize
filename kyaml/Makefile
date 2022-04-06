# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

include ../Makefile-tools.mk

.PHONY: generate fix vet fmt test lint tidy clean

$(MYGOBIN)/k8scopy:
	( cd ../cmd/k8scopy; go install . )

all: generate fix vet fmt test lint tidy

k8sGenDir := yaml/internal/k8sgen/pkg

generate: $(MYGOBIN)/stringer $(MYGOBIN)/k8scopy
	go generate ./...

clean:
	rm -rf $(k8sGenDir)

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  run ./... \
	  --path-prefix=kyaml \
	  -c ../.golangci.yml \
	  --skip-dirs yaml/internal/k8sgen/pkg \
	  --skip-dirs internal/forked

test:
	go test -v -cover ./...

fix:
	go fix ./...

fmt:
	go fmt $(shell go list ./... | grep -v "/kyaml/internal/forked/github.com/go-yaml/yaml")

tidy:
	go mod tidy

vet:
	go vet $(shell go list ./... | grep -v "/kyaml/internal/forked/github.com/go-yaml/yaml")

build:
	go build -v ./...
