# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for kustomize CLI and API.

MYGOBIN := $(shell go env GOPATH)/bin
SHELL := /bin/bash
export PATH := $(MYGOBIN):$(PATH)

.PHONY: all
all: verify-kustomize

.PHONY: verify-kustomize
verify-kustomize: \
	lint-kustomize \
	test-unit-kustomize-all \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-3.8.0 \
	test-examples-kustomize-against-3.8.2

# The following target referenced by a file in
# https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/kustomize
.PHONY: prow-presubmit-check
prow-presubmit-check: \
	lint-kustomize \
	test-unit-kustomize-all \
	test-unit-cmd-all \
	test-go-mod \
	test-examples-kustomize-against-HEAD \
	test-examples-kustomize-against-3.8.0 \
	test-examples-kustomize-against-3.8.2

.PHONY: verify-kustomize-e2e
verify-kustomize-e2e: test-examples-e2e-kustomize

# Other builds in this repo might want a different linter version.
# Without one Makefile to rule them all, the different makes
# cannot assume that golanci-lint is at the version they want
# since everything uses the same implicit GOPATH.
# This installs in a temp dir to avoid overwriting someone else's
# linter, then installs in MYGOBIN with a new name.
# Version pinned by hack/go.mod
$(MYGOBIN)/golangci-lint-kustomize:
	( \
		set -e; \
		cd hack; \
		GO111MODULE=on go build -tags=tools -o $(MYGOBIN)/golangci-lint-kustomize github.com/golangci/golangci-lint/cmd/golangci-lint; \
	)

$(MYGOBIN)/gorepomod:
	cd api; \
	go install github.com/monopole/gorepomod

$(MYGOBIN)/mdrip:
	cd api; \
	go install github.com/monopole/mdrip

$(MYGOBIN)/stringer:
	cd api; \
	go install golang.org/x/tools/cmd/stringer

$(MYGOBIN)/goimports:
	cd api; \
	go install golang.org/x/tools/cmd/goimports

# Install resource from whatever is checked out.
$(MYGOBIN)/resource:
	cd cmd/resource; \
	go install .

# To pin pluginator, use this recipe instead:
#   cd api;
#   go install sigs.k8s.io/kustomize/pluginator/v2
$(MYGOBIN)/pluginator:
	cd pluginator; \
	go install .

# Install kustomize from whatever is checked out.
$(MYGOBIN)/kustomize:
	cd kustomize; \
	go install .

.PHONY: install-tools
install-tools: \
	$(MYGOBIN)/goimports \
	$(MYGOBIN)/golangci-lint-kustomize \
	$(MYGOBIN)/gh \
	$(MYGOBIN)/gorepomod \
	$(MYGOBIN)/mdrip \
	$(MYGOBIN)/pluginator \
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
	HashTransformer.go \
	ImageTagTransformer.go \
	LabelTransformer.go \
	LegacyOrderTransformer.go \
	NamespaceTransformer.go \
	PatchJson6902Transformer.go \
	PatchStrategicMergeTransformer.go \
	PatchTransformer.go \
	PrefixSuffixTransformer.go \
	ReplicaCountTransformer.go \
	SecretGenerator.go \
	ValueAddTransformer.go

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
$(pGen)/HashTransformer.go: $(pSrc)/hashtransformer/HashTransformer.go
$(pGen)/ImageTagTransformer.go: $(pSrc)/imagetagtransformer/ImageTagTransformer.go
$(pGen)/LabelTransformer.go: $(pSrc)/labeltransformer/LabelTransformer.go
$(pGen)/LegacyOrderTransformer.go: $(pSrc)/legacyordertransformer/LegacyOrderTransformer.go
$(pGen)/NamespaceTransformer.go: $(pSrc)/namespacetransformer/NamespaceTransformer.go
$(pGen)/PatchJson6902Transformer.go: $(pSrc)/patchjson6902transformer/PatchJson6902Transformer.go
$(pGen)/PatchStrategicMergeTransformer.go: $(pSrc)/patchstrategicmergetransformer/PatchStrategicMergeTransformer.go
$(pGen)/PatchTransformer.go: $(pSrc)/patchtransformer/PatchTransformer.go
$(pGen)/PrefixSuffixTransformer.go: $(pSrc)/prefixsuffixtransformer/PrefixSuffixTransformer.go
$(pGen)/ReplicaCountTransformer.go: $(pSrc)/replicacounttransformer/ReplicaCountTransformer.go
$(pGen)/SecretGenerator.go: $(pSrc)/secretgenerator/SecretGenerator.go
$(pGen)/ValueAddTransformer.go: $(pSrc)/valueaddtransformer/ValueAddTransformer.go

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

.PHONY: kustomize-external-go-plugin-build
kustomize-external-go-plugin-build:
	./hack/buildExternalGoPlugins.sh ./plugin

.PHONY: kustomize-external-go-plugin-clean
kustomize-external-go-plugin-clean:
	./hack/buildExternalGoPlugins.sh ./plugin clean

### End kustomize plugin rules.

.PHONY: lint-kustomize
lint-kustomize: install-tools $(builtinplugins)
	cd api; \
	$(MYGOBIN)/golangci-lint-kustomize -c ../.golangci-kustomize.yml run ./...
	cd kustomize; \
	$(MYGOBIN)/golangci-lint-kustomize -c ../.golangci-kustomize.yml run ./...
	cd pluginator; \
	$(MYGOBIN)/golangci-lint-kustomize -c ../.golangci-kustomize.yml run ./...

# Used to add non-default compilation flags when experimenting with
# plugin-to-api compatibility checks.
.PHONY: build-kustomize-api
build-kustomize-api: $(builtinplugins)
	cd api; go build ./...

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
	./travis/kyaml-pre-commit.sh

test-go-mod:
	./travis/check-go-mod.sh

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
test-examples-kustomize-against-3.8.0: $(MYGOBIN)/mdrip
	( \
		set -e; \
		tag=v3.8.0; \
		/bin/rm -f $(MYGOBIN)/kustomize; \
		echo "Installing kustomize $$tag."; \
		GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@$${tag}; \
		./hack/testExamplesAgainstKustomize.sh $$tag; \
		echo "Reinstalling kustomize from HEAD."; \
		cd kustomize; go install .; \
	)

.PHONY:
test-examples-kustomize-against-3.8.2: $(MYGOBIN)/mdrip
	( \
		set -e; \
		tag=v3.8.2; \
		/bin/rm -f $(MYGOBIN)/kustomize; \
		echo "Installing kustomize $$tag."; \
		GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@$${tag}; \
		./hack/testExamplesAgainstKustomize.sh $$tag; \
		echo "Reinstalling kustomize from HEAD."; \
		cd kustomize; go install .; \
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
		wget https://github.com/instrumenta/kubeval/releases/latest/download/kubeval-linux-amd64.tar.gz; \
		tar xf kubeval-linux-amd64.tar.gz; \
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
		tgzFile=helm-v2.13.1-linux-amd64.tar.gz; \
		wget https://storage.googleapis.com/kubernetes-helm/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv linux-amd64/helm $(MYGOBIN)/helmV2; \
		rm -rf $$d \
	)

# Helm V3 differs from helm V2; downloading it to provide coverage for the
# chart inflator plugin under helm v3.
$(MYGOBIN)/helmV3:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=helm-v3.2.0-rc.1-linux-amd64.tar.gz; \
		wget https://get.helm.sh/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv linux-amd64/helm $(MYGOBIN)/helmV3; \
		rm -rf $$d \
	)

# Default version of helm is v2 for the time being.
$(MYGOBIN)/helm: $(MYGOBIN)/helmV2
	ln -s $(MYGOBIN)/helmV2 $(MYGOBIN)/helm

$(MYGOBIN)/kind:
	( \
        set -e; \
        d=$(shell mktemp -d); cd $$d; \
        wget -O ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(shell uname)-amd64; \
        chmod +x ./kind; \
        mv ./kind $(MYGOBIN); \
        rm -rf $$d; \
	)

# linux only.
$(MYGOBIN)/gh:
	( \
		set -e; \
		d=$(shell mktemp -d); cd $$d; \
		tgzFile=gh_1.0.0_linux_amd64.tar.gz; \
		wget https://github.com/cli/cli/releases/download/v1.0.0/$$tgzFile; \
		tar -xvzf $$tgzFile; \
		mv gh_1.0.0_linux_amd64/bin/gh  $(MYGOBIN)/gh; \
		rm -rf $$d \
	)

.PHONY: clean
clean: kustomize-external-go-plugin-clean
	go clean --cache
	rm -f $(builtinplugins)
	rm -f $(MYGOBIN)/pluginator
	rm -f $(MYGOBIN)/kustomize
	rm -f $(MYGOBIN)/golangci-lint-kustomize

# Nuke the site from orbit.  It's the only way to be sure.
.PHONY: nuke
nuke: clean
	go clean --modcache
