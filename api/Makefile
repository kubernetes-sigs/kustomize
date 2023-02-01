# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

include ../Makefile-modules.mk

test:
	go test -v -timeout 45m -cover ./... -ldflags "-X sigs.k8s.io/kustomize/api/provenance.buildDate=2023-01-31T23:38:41Z -X sigs.k8s.io/kustomize/api/provenance.version=(test)"

build:
	go build -ldflags "-X sigs.k8s.io/kustomize/api/provenance.buildDate=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")" ./...

generate: $(MYGOBIN)/k8scopy $(MYGOBIN)/stringer
	go generate ./...
