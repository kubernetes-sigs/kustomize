# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for the kustomize repo.

MYGOBIN := $(shell go env GOPATH)/bin
PATH := $(PATH):$(MYGOBIN)
SHELL := env PATH="$(PATH)" /bin/bash

.PHONY: all
all: verify-kustomize

.PHONY: verify-kustomize
verify-kustomize: \
	lint-kustomize \
	test-unit-kustomize-all \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-latest

# Version pinned by api/go.mod
$(MYGOBIN)/golangci-lint:
	cd api; \
	go install github.com/golangci/golangci-lint/cmd/golangci-lint

# Version pinned by api/go.mod
$(MYGOBIN)/mdrip:
	cd api; \
	go install github.com/monopole/mdrip

# Version pinned by api/go.mod
$(MYGOBIN)/stringer:
	cd api; \
	go install golang.org/x/tools/cmd/stringer

# Version pinned by api/go.mod
$(MYGOBIN)/goimports:
	cd api; \
	go install golang.org/x/tools/cmd/goimports

# Version pinned by api/go.mod
$(MYGOBIN)/pluginator:
	cd api; \
	go install sigs.k8s.io/kustomize/pluginator/v2

# Install kustomize from whatever is checked out.
$(MYGOBIN)/kustomize:
	cd kustomize; \
	go install .

.PHONY: install-tools
install-tools: \
	$(MYGOBIN)/goimports \
	$(MYGOBIN)/golangci-lint \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/pluginator \
	$(MYGOBIN)/stringer

# Builtin plugins are generated code.
# Add new items here to create new builtins.
builtinplugins = \
	api/builtins/annotationstransformer.go \
	api/builtins/configmapgenerator.go \
	api/builtins/hashtransformer.go \
	api/builtins/imagetagtransformer.go \
	api/builtins/inventorytransformer.go \
	api/builtins/labeltransformer.go \
	api/builtins/legacyordertransformer.go \
	api/builtins/namespacetransformer.go \
	api/builtins/patchjson6902transformer.go \
	api/builtins/patchstrategicmergetransformer.go \
	api/builtins/patchtransformer.go \
	api/builtins/prefixsuffixtransformer.go \
	api/builtins/replicacounttransformer.go \
	api/builtins/secretgenerator.go

.PHONY: lint-kustomize
lint-kustomize: install-tools $(builtinplugins)
	cd api; $(MYGOBIN)/golangci-lint run ./...
	cd kustomize; $(MYGOBIN)/golangci-lint run ./...
	cd pluginator; $(MYGOBIN)/golangci-lint run ./...

api/builtins/%.go: $(MYGOBIN)/pluginator
	@echo "generating $*"; \
	cd plugin/builtin/$*; \
	go generate .; \
	cd ../../../api/builtins; \
	$(MYGOBIN)/goimports -w $*.go

.PHONY: test-unit-kustomize-api
test-unit-kustomize-api: $(builtinplugins)
	cd api; go test ./...

.PHONY: test-unit-kustomize-plugins
test-unit-kustomize-plugins:
	./hack/testUnitKustomizePlugins.sh

.PHONY: test-unit-kustomize-cli
test-unit-kustomize-cli:
	cd kustomize; go test ./...

.PHONY: test-unit-kustomize-all
test-unit-kustomize-all: \
	test-unit-kustomize-api \
	test-unit-kustomize-cli \
	test-unit-kustomize-plugins

.PHONY:
test-examples-kustomize-against-HEAD: $(MYGOBIN)/kustomize $(MYGOBIN)/mdrip
	./hack/testExamplesAgainstKustomize.sh HEAD

.PHONY:
test-examples-kustomize-against-latest: $(MYGOBIN)/mdrip
	/bin/rm -f $(MYGOBIN)/kustomize; \
	echo "Installing kustomize from latest."; \
	go install sigs.k8s.io/kustomize/kustomize/v3; \
	./hack/testExamplesAgainstKustomize.sh latest; \
	echo "Reinstalling kustomize from HEAD."; \
	cd kustomize; go install .

# linux only.
# This is for testing an example plugin that
# uses kubeval for validation.
# Don't want to add a hard dependence in go.mod file
# to github.com/instrumenta/kubeval.
# Instead, download the binary.
$(MYGOBIN)/kubeval:
	d=$(shell mktemp -d); cd $$d; \
	wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-linux-amd64.tar.gz; \
	tar xf kubeval-linux-amd64.tar.gz; \
	mv kubeval $(MYGOBIN); \
	rm -rf $$d

# linux only.
# This is for testing an example plugin that
# uses helm to inflate a chart for subsequent kustomization.
# Don't want to add a hard dependence in go.mod file
# to helm.
# Instead, download the binary.
$(MYGOBIN)/helm:
	d=$(shell mktemp -d); cd $$d; \
	wget https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz; \
	tar -xvzf helm-v2.13.1-linux-amd64.tar.gz; \
	mv linux-amd64/helm $(MYGOBIN); \
	rm -rf $$d

.PHONY: clean
clean:
	rm -f $(builtinplugins)
	rm -f $(MYGOBIN)/pluginator
	rm -f $(MYGOBIN)/kustomize

.PHONY: nuke
nuke: clean
	sudo rm -rf $(shell go env GOPATH)/pkg/mod/sigs.k8s.io
