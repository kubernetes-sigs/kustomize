# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# This example uses a pinned version of the kustomize image, as an example of
# what users should do downstream, to ensure supply chain security.
#
# Whenever KUSTOMIZE_IMAGE_TAG & KUSTOMIZE_IMAGE_SHA are updated,
# EXAMPLE_IMAGE_TAG should also be updated to a new patch version.
# For reference, see
# https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-kustomize/images.yaml
EXAMPLE_IMAGE_TAG=v0.1.1
KUSTOMIZE_IMAGE_TAG=v5.4.1
KUSTOMIZE_IMAGE_SHA=sha256:7492c35d6fbe64e05100009915167a37b285ca7391067fa0c7bec9a7d1856882

all:
	true

.PHONY: image-build
image-build:
	docker build image \
		--build-arg "KUSTOMIZE_IMAGE_TAG=$(KUSTOMIZE_IMAGE_TAG)" \
		--build-arg "KUSTOMIZE_IMAGE_SHA=$(KUSTOMIZE_IMAGE_SHA)" \
		-t gcr.io/kustomize-functions/example-cockroachdb:$(EXAMPLE_IMAGE_TAG)

.PHONY: image
image: image-build
	docker push gcr.io/kustomize-functions/example-cockroachdb:$(EXAMPLE_IMAGE_TAG)
