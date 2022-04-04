# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: generate fix vet fmt test build tidy

include ../../../../../../Makefile-modules.mk

image:
	docker build . -t gcr.io/kustomize-functions/e2econtainerconfig
	docker push gcr.io/kustomize-functions/e2econtainerconfig
