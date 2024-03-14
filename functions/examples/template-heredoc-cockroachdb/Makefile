# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: image, kustomize-digest

FUNCTION_VERSION=v0.1.1
FUNCTION_IMAGE=gcr.io/kustomize-functions/example-cockroachdb:$(FUNCTION_VERSION)

# Update KUSTOMIZE_IMAGE_DIGEST & FUNCTION_VERSION whenever KUSTOMIZE_VERSION is updated!
# Use `make kustomize-digest` after changing KUSTOMIZE_VERSION to discover the
# new value to use for KUSTOMIZE_IMAGE_DIGEST.
KUSTOMIZE_VERSION=v5.3.0
KUSTOMIZE_IMAGE_NAME=registry.k8s.io/kustomize/kustomize
KUSTOMIZE_IMAGE_DIGEST=sha256:74b25b7e6d497336b9f89d152c9770a8318716ea6fe8962791598d195cb11488
KUSTOMIZE_IMAGE=$(KUSTOMIZE_IMAGE_NAME):$(KUSTOMIZE_VERSION)@$(KUSTOMIZE_IMAGE_DIGEST)
# To ensure the image has not changed, pin to a specific checksum for this version.

all:
	true

# Print the current digest for the pinned kustomize image version.
kustomize-digest:
	@docker pull $(KUSTOMIZE_IMAGE_NAME):$(KUSTOMIZE_VERSION) \
		| grep "Digest" | cut -d' ' -f2

# Build and push the function image
image:
	docker build image \
		--build-arg KUSTOMIZE_VERSION=$(KUSTOMIZE_VERSION) \
		-t $(FUNCTION_IMAGE)
	docker push $(FUNCTION_IMAGE)
