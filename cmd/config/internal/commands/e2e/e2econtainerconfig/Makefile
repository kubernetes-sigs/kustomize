# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate license fix vet fmt test build tidy

GOBIN := $(shell go env GOPATH)/bin

test:
	go test

image:
	docker build . -t gcr.io/kustomize-functions/e2econtainerconfig
	docker push gcr.io/kustomize-functions/e2econtainerconfig
