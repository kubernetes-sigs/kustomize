# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
GOHOSTOS = $(shell go env GOHOSTOS)
MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

REPO_ROOT := $(shell git rev-parse --show-toplevel)
TOOLS_DIR ?= $(REPO_ROOT)/.bin

GOLANGCI_LINT_DEFAULT := $(TOOLS_DIR)/golangci-lint$(if $(filter windows,$(GOHOSTOS)),.exe)

# Set GOLANGCI_LINT to use a pre-installed binary and skip the download, for
# example: make lint GOLANGCI_LINT=golangci-lint
ifeq ($(origin GOLANGCI_LINT), undefined)
GOLANGCI_LINT := $(GOLANGCI_LINT_DEFAULT)
GOLANGCI_LINT_PREREQUISITE := ensure-golangci-lint
endif
export GOLANGCI_LINT

# determines whether to run tests that only behave locally; can be overridden by override variable
export IS_LOCAL = false

.PHONY: install-out-of-tree-tools
install-out-of-tree-tools: \
	$(MYGOBIN)/goimports \
	$(GOLANGCI_LINT_PREREQUISITE) \
	$(MYGOBIN)/helmV3 \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/stringer

.PHONY: uninstall-out-of-tree-tools
uninstall-out-of-tree-tools:
	rm -f $(MYGOBIN)/goimports
	rm -f "$(GOLANGCI_LINT_DEFAULT)"
	rm -f $(MYGOBIN)/helmV3
	rm -f $(MYGOBIN)/mdrip
	rm -f $(MYGOBIN)/stringer

.PHONY: ensure-golangci-lint
ensure-golangci-lint:
	"$(REPO_ROOT)/hack/ensure-golangci-lint.sh" \
		-b "$(TOOLS_DIR)" \
		"$(shell cat "$(REPO_ROOT)/.golangci-lint-version")"

.PHONY: $(MYGOBIN)/mdrip
$(MYGOBIN)/mdrip:
	cd $(REPO_ROOT)/hack && go install github.com/monopole/mdrip

.PHONY: $(MYGOBIN)/stringer
$(MYGOBIN)/stringer:
	cd $(REPO_ROOT)/hack && go install golang.org/x/tools/cmd/stringer

.PHONY: $(MYGOBIN)/goimports
$(MYGOBIN)/goimports:
	cd $(REPO_ROOT)/hack && go install golang.org/x/tools/cmd/goimports

.PHONY: $(MYGOBIN)/mdtogo
$(MYGOBIN)/mdtogo:
	cd $(REPO_ROOT)/hack && go install sigs.k8s.io/kustomize/cmd/mdtogo

.PHONY: $(MYGOBIN)/addlicense
$(MYGOBIN)/addlicense:
	cd $(REPO_ROOT)/hack && go install github.com/google/addlicense

.PHONY: $(MYGOBIN)/kind
$(MYGOBIN)/kind:
	cd $(REPO_ROOT)/hack && go install sigs.k8s.io/kind

.PHONY: $(MYGOBIN)/controller-gen
$(MYGOBIN)/controller-gen:
	cd $(REPO_ROOT)/hack && go install sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: $(MYGOBIN)/embedmd
$(MYGOBIN)/embedmd:
	cd $(REPO_ROOT)/hack && go install github.com/campoy/embedmd

.PHONY: $(MYGOBIN)/go-apidiff
$(MYGOBIN)/go-apidiff:
	cd $(REPO_ROOT)/hack && go install github.com/joelanford/go-apidiff

.PHONY: $(MYGOBIN)/gh
$(MYGOBIN)/gh:
	cd $(REPO_ROOT)/hack && go install github.com/cli/cli/cmd/gh

.PHONY: $(MYGOBIN)/kubeval
$(MYGOBIN)/kubeval:
	cd $(REPO_ROOT)/hack && go install github.com/instrumenta/kubeval

# Helm V3 differs from helm V2; downloading it to provide coverage for the
# chart inflator plugin under helm v3.
.PHONY: $(MYGOBIN)/helmV3
$(MYGOBIN)/helmV3:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=helm-v3.10.2-$(GOOS)-$(GOARCH).tar.gz; \
		wget https://get.helm.sh/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv $(GOOS)-$(GOARCH)/helm $(MYGOBIN)/helmV3; \
		rm -rf $$d \
	)
