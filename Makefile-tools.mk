# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

GOLANGCI_LINT_VERSION=v1.50.1

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

# determines whether to run tests that only behave locally; can be overridden by override variable
export IS_LOCAL = false

.PHONY: install-out-of-tree-tools
install-out-of-tree-tools: \
	$(MYGOBIN)/goimports \
	$(MYGOBIN)/golangci-lint \
	$(MYGOBIN)/helmV3 \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/stringer \
	$(MYGOBIN)/goimports

.PHONY: uninstall-out-of-tree-tools
uninstall-out-of-tree-tools:
	rm -f $(MYGOBIN)/goimports
	rm -f $(MYGOBIN)/golangci-lint
	rm -f $(MYGOBIN)/helmV3
	rm -f $(MYGOBIN)/mdrip
	rm -f $(MYGOBIN)/stringer

$(MYGOBIN)/golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(MYGOBIN)/mdrip:
	go install github.com/monopole/mdrip@v1.0.2

$(MYGOBIN)/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

$(MYGOBIN)/goimports:
	go install golang.org/x/tools/cmd/goimports@latest

$(MYGOBIN)/mdtogo:
	go install sigs.k8s.io/kustomize/cmd/mdtogo@latest

$(MYGOBIN)/addlicense:
	go install github.com/google/addlicense@latest

$(MYGOBIN)/statik:
	go install github.com/rakyll/statik@latest

$(MYGOBIN)/goreleaser:
	go install github.com/goreleaser/goreleaser@v0.179.0 # https://github.com/kubernetes-sigs/kustomize/issues/4542

$(MYGOBIN)/kind:
	( \
        set -e; \
        d=$(shell mktemp -d); cd $$d; \
        wget -O ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(GOOS)-$(GOARCH); \
        chmod +x ./kind; \
        mv ./kind $(MYGOBIN); \
        rm -rf $$d; \
	)

# linux only.
$(MYGOBIN)/gh:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=gh_1.0.0_$(GOOS)_$(GOARCH).tar.gz; \
		wget https://github.com/cli/cli/releases/download/v1.0.0/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv gh_1.0.0_$(GOOS)_$(GOARCH)/bin/gh  $(MYGOBIN)/gh; \
		rm -rf $$d \
	)

# linux only.
# This is for testing an example plugin that
# uses kubeval for validation.
# Don't want to add a hard dependence in go.mod file
# to github.com/instrumenta/kubeval.
# Instead, download the binary.
$(MYGOBIN)/kubeval:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-$(GOOS)-$(GOARCH).tar.gz; \
		tar xf kubeval-$(GOOS)-$(GOARCH).tar.gz; \
		mv kubeval $(MYGOBIN); \
		rm -rf $$d; \
	)

# Helm V3 differs from helm V2; downloading it to provide coverage for the
# chart inflator plugin under helm v3.
$(MYGOBIN)/helmV3:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
<<<<<<< HEAD
		tgzFile=helm-v3.10.2-$(GOOS)-$(GOARCH).tar.gz; \
=======
		tgzFile=helm-v3.8.2-$(GOOS)-$(GOARCH).tar.gz; \
>>>>>>> bffa1db93 (updated to helm v3.8+)
		wget https://get.helm.sh/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv $(GOOS)-$(GOARCH)/helm $(MYGOBIN)/helmV3; \
		rm -rf $$d \
	)
