# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

include ../../../../../../Makefile-modules.mk

.PHONY: image
image:
	docker build . -t gcr.io/kustomize-functions/e2econtainerconfig
	docker push gcr.io/kustomize-functions/e2econtainerconfig
