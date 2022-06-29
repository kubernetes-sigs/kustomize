# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

include ../Makefile-examples-base.mk

.PHONY: image
image:
	docker build image -t gcr.io/kustomize-functions/example-tshirt:v0.3.0
	docker push gcr.io/kustomize-functions/example-tshirt:v0.3.0
