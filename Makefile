# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for kustomize CLI and API.

LATEST_V4_RELEASE=v4.5.4

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
install-tools: \
	install-local-tools \
	install-out-of-tree-tools

.PHONY: uninstall-tools
uninstall-tools: \
	uninstall-local-tools \
	uninstall-out-of-tree-tools

.PHONY: install-local-tools
install-local-tools: \
	$(MYGOBIN)/gorepomod \
	$(MYGOBIN)/k8scopy \
	$(MYGOBIN)/pluginator

.PHONY: uninstall-local-tools
uninstall-local-tools:
	rm -f $(MYGOBIN)/gorepomod
	rm -f $(MYGOBIN)/k8scopy
	rm -f $(MYGOBIN)/pluginator

# Build from local source.
$(MYGOBIN)/gorepomod:
	cd cmd/gorepomod; \
	go install .

# Build from local source.
$(MYGOBIN)/k8scopy:
	cd cmd/k8scopy; \
	go install .

# Build from local source.
$(MYGOBIN)/pluginator:
	cd cmd/pluginator; \
	go install .


# --- Build targets ---

# Build from local source.
$(MYGOBIN)/kustomize: build-kustomize-api
	cd kustomize; \
	go install .

kustomize: $(MYGOBIN)/kustomize

# Used to add non-default compilation flags when experimenting with
# plugin-to-api compatibility checks.
.PHONY: build-kustomize-api
build-kustomize-api: $(MYGOBIN)/goimports $(builtinplugins)
	cd api; $(MAKE) build

.PHONY: generate-kustomize-api
generate-kustomize-api:
	cd api; $(MAKE) generate


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
	test-examples-kustomize-against-v4-release

# The following target referenced by a file in
# https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/kustomize
.PHONY: prow-presubmit-check
prow-presubmit-check: \
	install-tools \
	test-unit-kustomize-plugins \
	test-go-mod \
	build-non-plugin-all \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-v4-release

.PHONY: license
license: $(MYGOBIN)/addlicense
	./hack/add-license.sh run

.PHONY: check-license
check-license: $(MYGOBIN)/addlicense
	./hack/add-license.sh check

.PHONY: lint
lint: $(MYGOBIN)/golangci-lint $(MYGOBIN)/goimports $(builtinplugins)
	./hack/for-each-module.sh "make lint"

.PHONY: test-unit-all
test-unit-all: \
	test-unit-non-plugin \
	test-unit-kustomize-plugins

# This target is used by our Github Actions CI to run unit tests for all non-plugin modules in multiple GOOS environments.
.PHONY: test-unit-non-plugin
test-unit-non-plugin: kustomize
	./hack/for-each-module.sh "make test" "./plugin/*" 15

.PHONY: build-non-plugin-all
build-non-plugin-all:
	./hack/for-each-module.sh "make build" "./plugin/*" 15

.PHONY: test-unit-kustomize-plugins
test-unit-kustomize-plugins:
	./hack/testUnitKustomizePlugins.sh

.PHONY: functions-examples-all
functions-examples-all:
	for dir in $(abspath $(wildcard functions/examples/*/.)); do \
		echo -e "\n---Running make tasks for function $$dir---"; \
		set -e; \
		cd $$dir; $(MAKE) all; \
	done

test-go-mod:
	./hack/for-each-module.sh "go list -m -json all > /dev/null && go mod tidy -v"

.PHONY:
verify-kustomize-e2e: $(MYGOBIN)/mdrip $(MYGOBIN)/kind
	( \
		set -e; \
		/bin/rm -f $(MYGOBIN)/kustomize; \
		echo "Installing kustomize from ."; \
		cd kustomize; go install .; cd ..; \
		./hack/testExamplesE2EAgainstKustomize.sh .; \
	)

.PHONY:
test-examples-kustomize-against-HEAD: $(MYGOBIN)/kustomize $(MYGOBIN)/mdrip
	./hack/testExamplesAgainstKustomize.sh HEAD

.PHONY:
test-examples-kustomize-against-v4-release: $(MYGOBIN)/mdrip
	./hack/testExamplesAgainstKustomize.sh v4@$(LATEST_V4_RELEASE)


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
