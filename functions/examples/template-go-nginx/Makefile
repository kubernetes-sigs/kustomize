# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate fix vet fmt test build tidy image

export KUSTOMIZE_ROOT ?= $(shell pwd | sed -E 's|(.*\/kustomize)/(.*)|\1|')
include $(KUSTOMIZE_ROOT)/Makefile-tools.mk

build:
	(cd image && go build -v -o $(GOBIN)/config-function .)

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

image:
	docker build image -t gcr.io/kustomize-functions/example-nginx:v0.2.0
	docker push gcr.io/kustomize-functions/example-nginx:v0.2.0
