# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

include ../Makefile-modules.mk

test:
	go test -v -timeout 45m -cover ./... -ldflags "-X sigs.k8s.io/kustomize/api/provenance.buildDate=2023-01-31T23:38:41Z -X sigs.k8s.io/kustomize/api/provenance.version=(test)"

build:
	go build -ldflags "-X sigs.k8s.io/kustomize/api/provenance.buildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")" ./...

generate: $(MYGOBIN)/k8scopy $(MYGOBIN)/stringer
	go generate ./...

lint: lint-api-static

## lint-api-static runs the linter on the API module
## with build-tag kustomize_disable_go_plugin_support
## this aims to catch any issues with the API module
## that would prevent the API module from being used
## when the go plugin support is disabled.
lint-api-static:
	$(MYGOBIN)/golangci-lint cache clean # Workaround for https://github.com/golangci/golangci-lint/issues/3228
	$(MYGOBIN)/golangci-lint \
	  -c $$KUSTOMIZE_ROOT/.golangci.yml \
	  --build-tags kustomize_disable_go_plugin_support \
	  --path-prefix api \
	  run ./...
