# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

ARG KUSTOMIZE_IMAGE_TAG
ARG KUSTOMIZE_IMAGE_SHA

FROM registry.k8s.io/kustomize/kustomize:${KUSTOMIZE_IMAGE_TAG}@${KUSTOMIZE_IMAGE_SHA}
RUN apk add --no-cache bash
COPY cockroachdb-template.sh /usr/local/bin/config-function
CMD ["config-function"]
