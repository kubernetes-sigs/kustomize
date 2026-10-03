# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for kustomize CLI and API.

LATEST_RELEASE=v5.8.2

SHELL := /usr/bin/env bash
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)

# Provide defaults for REPO_OWNER and REPO_NAME if not present.
# Typically these values would be provided by Prow.
ifndef REPO_OWNER
REPO_OWNER := "kubernetes-sigs"
endif

ifndef REPO_NAME
REPO_NAME := "kustomize"
endif


# --- Plugins ---
include Makefile-plugins.mk


# --- Tool management ---
include Makefile-tools.mk

.PHONY: install-tools
install-tools: install-out-of-tree-tools

.PHONY: uninstall-tools
uninstall-tools: uninstall-out-of-tree-tools


# --- Build targets ---

# Build from local source.
$(MYGOBIN)/kustomize: build-kustomize-api
	cd kustomize && go install -ldflags \
	"-X sigs.k8s.io/kustomize/api/provenance.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
	 -X sigs.k8s.io/kustomize/api/provenance.version=$(shell git describe --tags --always --dirty)" \
	.

kustomize: $(MYGOBIN)/kustomize

# Used to add non-default compilation flags when experimenting with
# plugin-to-api compatibility checks.
.PHONY: build-kustomize-api
build-kustomize-api: $(builtinplugins)
	cd api && $(MAKE) build

.PHONY: generate-kustomize-api
generate-kustomize-api:
	cd api && $(MAKE) generate


# --- Verification targets ---
.PHONY: verify-kustomize-repo
verify-kustomize-repo: \
	install-tools \
	lint \
	check-license \
	test-unit-all \
	build-non-plugin-all \
	test-go-mod \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-latest-release

# The following target referenced by a file in
# https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/kustomize
.PHONY: prow-presubmit-check
prow-presubmit-check: \
	install-tools \
	workspace-sync \
	generate-kustomize-builtin-plugins \
	builtin-plugins-diff \
	test-unit-kustomize-plugins \
	test-go-mod \
	build-non-plugin-all \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-latest-release

.PHONY: license
license:
	./hack/add-license.sh run

.PHONY: check-license
check-license:
	./hack/add-license.sh check

.PHONY: lint
lint: $(GOLANGCI_LINT_PREREQUISITE) $(builtinplugins)
	./hack/for-each-module.sh "make lint"

APIDIFF_BASE_REF ?= master

.PHONY: apidiff
apidiff: ## Run go-apidiff to verify API differences compared with APIDIFF_BASE_REF
	go tool go-apidiff "$(APIDIFF_BASE_REF)" --compare-imports --print-compatible --repo-path=.

.PHONY: test-unit-all
test-unit-all: \
	test-unit-non-plugin \
	test-unit-kustomize-plugins

.PHONY: test-unit-non-plugin
test-unit-non-plugin:
	./hack/for-each-module.sh "make test" "./plugin/*" 20

# This target is used by our Github Actions CI to run unit tests for all non-plugin and non-released modules in multiple GOOS environments.
.PHONY: test-unit-non-plugin-and-non-released
test-unit-non-plugin-and-non-released:
	./hack/for-each-module.sh "make test" "./plugin/*|./kyaml/go.mod|./cmd/config/go.mod|./api/go.mod|./kustomize/go.mod" 16

.PHONY: build-non-plugin-all
build-non-plugin-all:
	./hack/for-each-module.sh "make build" "./plugin/*" 20

.PHONY: test-unit-kustomize-plugins
test-unit-kustomize-plugins: build-kustomize-external-go-plugin
	./hack/testUnitKustomizePlugins.sh

.PHONY: functions-examples-all
functions-examples-all:
	for dir in $(abspath $(wildcard functions/examples/*/.)); do \
		echo -e "\n---Running make tasks for function $$dir---"; \
		set -e; \
		cd $$dir; $(MAKE) all; \
	done

test-go-mod:
	./hack/for-each-module.sh "go mod tidy -v"

.PHONY:
verify-kustomize-e2e: $(MYGOBIN)/kind
	( \
		set -e; \
		/bin/rm -f $(MYGOBIN)/kustomize; \
		echo "Installing kustomize from ."; \
		cd kustomize; go install .; cd ..; \
		./hack/testExamplesE2EAgainstKustomize.sh .; \
	)

.PHONY:
test-examples-kustomize-against-HEAD: $(MYGOBIN)/kustomize
	./hack/testExamplesAgainstKustomize.sh HEAD

.PHONY:
test-examples-kustomize-against-latest-release:
	./hack/testExamplesAgainstKustomize.sh v5@$(LATEST_RELEASE)

# Pushes dependencies in the go.work file back to go.mod files of each workspace module.
.PHONY: workspace-sync
workspace-sync:
	go work sync
	./hack/doGoMod.sh tidy
	# Record checksums needed to load the complete workspace module graph.
	go list -m all > /dev/null

# --- Cleanup targets ---
.PHONY: clean
clean: clean-kustomize-external-go-plugin uninstall-tools
	go clean --cache
	rm -f $(builtinplugins)
	rm -f $(MYGOBIN)/kustomize

# Nuke the site from orbit.  It's the only way to be sure.
.PHONY: nuke
nuke: clean
	go clean --modcache
