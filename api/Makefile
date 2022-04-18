# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

include ../Makefile-modules.mk

test:
	go test -v -timeout 45m -cover ./... -ldflags "-X sigs.k8s.io/kustomize/api/provenance.version=v444.333.222"
	cd krusty/openapitests; OPENAPI_TEST=true go test -v -timeout 45m -p 1 -cover ./...

build:
	go build -ldflags "-X sigs.k8s.io/kustomize/api/provenance.version=v444.333.222" ./...

generate: $(MYGOBIN)/k8scopy $(MYGOBIN)/stringer
	go generate ./...
