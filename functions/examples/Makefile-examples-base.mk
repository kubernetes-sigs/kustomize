# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

.PHONY: generate fix vet fmt test build tidy

export KUSTOMIZE_ROOT ?= $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\1|')
include $(KUSTOMIZE_ROOT)/Makefile-tools.mk

build:
	(cd image && go build -v -o $(MYGOBIN)/config-function .)

all: generate build fix vet fmt test lint tidy

fix:
	(cd image && go fix ./...)

fmt:
	(cd image && go fmt ./...)

generate: $(MYGOBIN)/mdtogo
	(cd image && GOBIN=$(MYGOBIN) go generate ./...)

tidy:
	(cd image && go mod tidy)

lint: $(MYGOBIN)/golangci-lint
	(cd image && $(MYGOBIN)/golangci-lint \
	  -c $$KUSTOMIZE_ROOT/.golangci.yml \
	  --path-prefix $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\2|') \
	  run ./...)

test:
	(cd image && go test -cover ./...)

vet:
	(cd image && go vet ./...)
