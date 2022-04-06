# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate fix vet fmt test build tidy clean

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

include ../../Makefile-tools.mk

$(MYGOBIN)/k8scopy:
	( cd ../k8scopy; go install . )

build:
	go build -v -o $(MYGOBIN)/kubectl-krm ./kubectl-krm

all: build fix vet fmt test lint tidy

k8sGenDir := internal/commands/internal/k8sgen/pkg

generate: $(MYGOBIN)/mdtogo $(MYGOBIN)/k8scopy
	GOBIN="${MYGOBIN}" go generate ./...

clean:
	rm -rf $(k8sGenDir)

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  --skip-dirs $(k8sGenDir) \
	  run ./... \
	  --path-prefix=cmd/config \
	  -c ../../.golangci.yml

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
