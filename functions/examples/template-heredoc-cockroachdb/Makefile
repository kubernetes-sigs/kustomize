# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: license image

GOBIN = $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN = $(shell go env GOPATH)/bin
endif

license:
	(which $(GOBIN)/addlicense || go get github.com/google/addlicense)
	$(GOBIN)/addlicense  -y 2022 -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

all: license

image:
	docker build image -t gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
	docker push gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
