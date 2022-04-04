# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

# only set this if not already set, so importing makefiles can override it
export KUSTOMIZE_ROOT ?= $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\1|')
include $(KUSTOMIZE_ROOT)/Makefile-tools.mk

.PHONY: lint test fix fmt tidy vet build

lint: $(MYGOBIN)/golangci-lint
	$(MYGOBIN)/golangci-lint \
	  -c $$KUSTOMIZE_ROOT/.golangci.yml \
	  --path-prefix $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\2|') \
	  run ./...

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

build:
	go build -v ./...
