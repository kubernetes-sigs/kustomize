# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for kustomize CLI and API.

SHELL := /usr/bin/env bash
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif
export PATH := $(MYGOBIN):$(PATH)
MODULES := '"cmd/config" "api/" "kustomize/" "kyaml/"'

# Provide defaults for REPO_OWNER and REPO_NAME if not present.
# Typically these values would be provided by Prow.
ifndef REPO_OWNER
REPO_OWNER := "kubernetes-sigs"
endif

ifndef REPO_NAME
REPO_NAME := "kustomize"
endif

.PHONY: all
all: install-tools verify-kustomize

.PHONY: verify-kustomize
verify-kustomize: \
	lint-kustomize \
	test-unit-kustomize-all \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-4.1

# The following target referenced by a file in
# https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/kustomize
.PHONY: prow-presubmit-check
prow-presubmit-check: \
	install-tools \
	lint-kustomize \
	test-multi-module \
	test-unit-kustomize-all \
	test-unit-cmd-all \
	test-go-mod \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-4.1

.PHONY: verify-kustomize-e2e
verify-kustomize-e2e: test-examples-e2e-kustomize

# Other builds in this repo might want a different linter version.
# Without one Makefile to rule them all, the different makes
# cannot assume that golanci-lint is at the version they want.
# This installs what kustomize wants to use.
$(MYGOBIN)/golangci-lint-kustomize:
	rm -f $(CURDIR)/hack/golangci-lint
	GOBIN=$(CURDIR)/hack go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.23.8
	mv $(CURDIR)/hack/golangci-lint $(MYGOBIN)/golangci-lint-kustomize

$(MYGOBIN)/mdrip:
	go install github.com/monopole/mdrip@v1.0.2

$(MYGOBIN)/stringer:
	go get golang.org/x/tools/cmd/stringer

$(MYGOBIN)/goimports:
	go get golang.org/x/tools/cmd/goimports

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

# Build from local source.
$(MYGOBIN)/prchecker:
	cd cmd/prchecker; \
	go install .

# Build from local source.
$(MYGOBIN)/kustomize: build-kustomize-api
	cd kustomize; \
	go install .

.PHONY: install-tools
install-tools: \
	$(MYGOBIN)/goimports \
	$(MYGOBIN)/golangci-lint-kustomize \
	$(MYGOBIN)/gorepomod \
	$(MYGOBIN)/helmV3 \
	$(MYGOBIN)/k8scopy \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/pluginator \
	$(MYGOBIN)/prchecker \
	$(MYGOBIN)/stringer

### Begin kustomize plugin rules.
#
# The rules to deal with builtin plugins are a bit
# complicated because
#
# - Every builtin plugin is a Go plugin -
#   meaning it gets its own module directory
#   (outside of the api module) with Go
#   code in a 'main' package per Go plugin rules.
# - kustomize locates plugins using the
#   'apiVersion' and 'kind' fields from the
#   plugin config file.
# - k8s wants CamelCase in 'kind' fields.
# - The module name (the last name in the path)
#   must be the lowercased 'kind' of the
#   plugin because Go and related tools
#   demand lowercase in import paths, but
#   allow CamelCase in file names.
# - the generated code must live in the api
#   module (it's linked into the api).

# Where all generated builtin plugin code should go.
pGen=api/builtins
# Where the builtin Go plugin modules live.
pSrc=plugin/builtin

_builtinplugins = \
	AnnotationsTransformer.go \
	ConfigMapGenerator.go \
	IAMPolicyGenerator.go \
	HashTransformer.go \
	ImageTagTransformer.go \
	LabelTransformer.go \
	LegacyOrderTransformer.go \
	NamespaceTransformer.go \
	PatchJson6902Transformer.go \
	PatchStrategicMergeTransformer.go \
	PatchTransformer.go \
	PrefixSuffixTransformer.go \
	ReplacementTransformer.go \
	ReplicaCountTransformer.go \
	SecretGenerator.go \
	ValueAddTransformer.go \
	HelmChartInflationGenerator.go

# Maintaining this explicit list of generated files, and
# adding it as a dependency to a few targets, to assure
# they get recreated if deleted.  The rules below on how
# to make them don't, by themselves, assure they will be
# recreated if deleted.
builtinplugins = $(patsubst %,$(pGen)/%,$(_builtinplugins))

# These rules are verbose, but assure that if a source file
# is modified, the corresponding generated file, and only
# that file, will be recreated.
$(pGen)/AnnotationsTransformer.go: $(pSrc)/annotationstransformer/AnnotationsTransformer.go
$(pGen)/ConfigMapGenerator.go: $(pSrc)/configmapgenerator/ConfigMapGenerator.go
$(pGen)/GkeSaGenerator.go: $(pSrc)/gkesagenerator/GkeSaGenerator.go
$(pGen)/HashTransformer.go: $(pSrc)/hashtransformer/HashTransformer.go
$(pGen)/ImageTagTransformer.go: $(pSrc)/imagetagtransformer/ImageTagTransformer.go
$(pGen)/LabelTransformer.go: $(pSrc)/labeltransformer/LabelTransformer.go
$(pGen)/LegacyOrderTransformer.go: $(pSrc)/legacyordertransformer/LegacyOrderTransformer.go
$(pGen)/NamespaceTransformer.go: $(pSrc)/namespacetransformer/NamespaceTransformer.go
$(pGen)/PatchJson6902Transformer.go: $(pSrc)/patchjson6902transformer/PatchJson6902Transformer.go
$(pGen)/PatchStrategicMergeTransformer.go: $(pSrc)/patchstrategicmergetransformer/PatchStrategicMergeTransformer.go
$(pGen)/PatchTransformer.go: $(pSrc)/patchtransformer/PatchTransformer.go
$(pGen)/PrefixSuffixTransformer.go: $(pSrc)/prefixsuffixtransformer/PrefixSuffixTransformer.go
$(pGen)/ReplacementTransformer.go: $(pSrc)/replacementtransformer/ReplacementTransformer.go
$(pGen)/ReplicaCountTransformer.go: $(pSrc)/replicacounttransformer/ReplicaCountTransformer.go
$(pGen)/SecretGenerator.go: $(pSrc)/secretgenerator/SecretGenerator.go
$(pGen)/ValueAddTransformer.go: $(pSrc)/valueaddtransformer/ValueAddTransformer.go
$(pGen)/HelmChartInflationGenerator.go: $(pSrc)/helmchartinflationgenerator/HelmChartInflationGenerator.go

# The (verbose but portable) Makefile way to convert to lowercase.
toLowerCase = $(subst A,a,$(subst B,b,$(subst C,c,$(subst D,d,$(subst E,e,$(subst F,f,$(subst G,g,$(subst H,h,$(subst I,i,$(subst J,j,$(subst K,k,$(subst L,l,$(subst M,m,$(subst N,n,$(subst O,o,$(subst P,p,$(subst Q,q,$(subst R,r,$(subst S,s,$(subst T,t,$(subst U,u,$(subst V,v,$(subst W,w,$(subst X,x,$(subst Y,y,$(subst Z,z,$1))))))))))))))))))))))))))

$(pGen)/%.go: $(MYGOBIN)/pluginator
	@echo "generating $*"
	( \
		set -e; \
		cd $(pSrc)/$(call toLowerCase,$*); \
		go generate .; \
		cd ../../../$(pGen); \
		$(MYGOBIN)/goimports -w $*.go \
	)

# Target is for debugging.
.PHONY: generate-kustomize-builtin-plugins
generate-kustomize-builtin-plugins: $(builtinplugins)

.PHONY: build-kustomize-external-go-plugin
build-kustomize-external-go-plugin:
	./hack/buildExternalGoPlugins.sh ./plugin

.PHONY: clean-kustomize-external-go-plugin
clean-kustomize-external-go-plugin:
	./hack/buildExternalGoPlugins.sh ./plugin clean

### End kustomize plugin rules.

.PHONY: lint-kustomize
lint-kustomize: $(MYGOBIN)/golangci-lint-kustomize $(builtinplugins)
	cd api; $(MYGOBIN)/golangci-lint-kustomize \
	  -c ../.golangci-kustomize.yml \
	  run ./...
	cd kustomize; $(MYGOBIN)/golangci-lint-kustomize \
	  -c ../.golangci-kustomize.yml \
	  run ./...
	cd cmd/pluginator; $(MYGOBIN)/golangci-lint-kustomize \
	  -c ../../.golangci-kustomize.yml \
	  run ./...

# Used to add non-default compilation flags when experimenting with
# plugin-to-api compatibility checks.
.PHONY: build-kustomize-api
build-kustomize-api: $(builtinplugins)
	cd api; go build ./...

.PHONY: generate-kustomize-api
generate-kustomize-api: $(MYGOBIN)/k8scopy
	cd api; go generate ./...

.PHONY: test-unit-kustomize-api
test-unit-kustomize-api: build-kustomize-api
	cd api; go test ./...  -ldflags "-X sigs.k8s.io/kustomize/api/provenance.version=v444.333.222"

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

test-unit-cmd-all:
	./hack/kyaml-pre-commit.sh

test-go-mod:
	./hack/check-go-mod.sh

# Environment variables are defined at
# https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#job-environment-variables
.PHONY: test-multi-module
test-multi-module: $(MYGOBIN)/prchecker
	( \
		export MYGOBIN=$(MYGOBIN); \
		export REPO_OWNER=$(REPO_OWNER); \
		export REPO_NAME=$(REPO_NAME); \
		export PULL_NUMBER=$(PULL_NUMBER); \
		export MODULES=$(MODULES); \
		./hack/check-multi-module.sh; \
	)

.PHONY:
test-examples-e2e-kustomize: $(MYGOBIN)/mdrip $(MYGOBIN)/kind
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
test-examples-kustomize-against-4.1: $(MYGOBIN)/mdrip
	./hack/testExamplesAgainstKustomize.sh v4@v4.1.2

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

# linux only.
# This is for testing an example plugin that uses helm to inflate a chart
# for subsequent kustomization.
# Don't want to add a hard dependence in go.mod file to helm.
# Instead, download the binaries.
$(MYGOBIN)/helmV2:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=helm-v2.13.1-$(GOOS)-$(GOARCH).tar.gz; \
		wget https://storage.googleapis.com/kubernetes-helm/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv $(GOOS)-$(GOARCH)/helm $(MYGOBIN)/helmV2; \
		rm -rf $$d \
	)

# Helm V3 differs from helm V2; downloading it to provide coverage for the
# chart inflator plugin under helm v3.
$(MYGOBIN)/helmV3:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=helm-v3.5.3-$(GOOS)-$(GOARCH).tar.gz; \
		wget https://get.helm.sh/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv $(GOOS)-$(GOARCH)/helm $(MYGOBIN)/helmV3; \
		rm -rf $$d \
	)

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

.PHONY: clean
clean: clean-kustomize-external-go-plugin
	go clean --cache
	rm -f $(builtinplugins)
	rm -f $(MYGOBIN)/goimports
	rm -f $(MYGOBIN)/golangci-lint-kustomize
	rm -f $(MYGOBIN)/kustomize
	rm -f $(MYGOBIN)/mdrip
	rm -f $(MYGOBIN)/prchecker
	rm -f $(MYGOBIN)/stringer

# Handle pluginator manually.
# rm -f $(MYGOBIN)/pluginator

# Nuke the site from orbit.  It's the only way to be sure.
.PHONY: nuke
nuke: clean
	go clean --modcache
