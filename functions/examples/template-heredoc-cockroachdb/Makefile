# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: image

all:
	true

image:
	docker build image -t gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
	docker push gcr.io/kustomize-functions/example-cockroachdb:v0.1.0
