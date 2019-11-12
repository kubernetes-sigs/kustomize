# This Makefile is (and must be) used by
# travis/pre-commit.sh to qualify pull requests.
#
# That script generates all the code that needs
# to be generated, and runs all the tests.
#
# Functionality in that script is gradually moving here.

MYGOBIN := $(shell go env GOPATH)/bin
PATH := $(PATH):$(MYGOBIN)
SHELL := env PATH=$(PATH) /bin/bash

.PHONY: all
all: pre-commit

# The pre-commit.sh script generates, lints and tests.
# It uses this makefile.  For more clarity, would like
# to stop that - any scripts invoked by targets here
# shouldn't "call back" to the makefile.
.PHONY: pre-commit
pre-commit:
	./travis/pre-commit.sh

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

.PHONY: lint
lint: install-tools $(builtinplugins)
	cd api; $(MYGOBIN)/golangci-lint run ./...
	cd kustomize; $(MYGOBIN)/golangci-lint run ./...
	cd pluginator; $(MYGOBIN)/golangci-lint run ./...

api/builtins/%.go: $(MYGOBIN)/pluginator
	@echo "generating $*"; \
	cd plugin/builtin/$*; \
	go generate .; \
	cd ../../../api/builtins; \
	$(MYGOBIN)/goimports -w $*.go

.PHONY: generate
generate: $(builtinplugins)

.PHONY: unit-test-api
unit-test-api: $(builtinplugins)
	cd api; go test ./...

.PHONY: unit-test-plugins
unit-test-plugins:
	./hack/runPluginUnitTests.sh

.PHONY: unit-test-kustomize
unit-test-kustomize:
	cd kustomize; go test ./...

.PHONY: unit-test-all
unit-test-all: unit-test-api unit-test-kustomize unit-test-plugins

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

.PHONY: nuke
nuke: clean
	sudo rm -rf $(shell go env GOPATH)/pkg/mod/sigs.k8s.io
