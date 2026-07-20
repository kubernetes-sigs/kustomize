# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

MYGOBIN = $(shell go env GOBIN)
ifeq ($(MYGOBIN),)
MYGOBIN = $(shell go env GOPATH)/bin
endif

# The kubernetes/kubernetes tag whose published OpenAPI document is the
# source of the precomputed strategic-merge patch metadata
# (zz_generated_patchmeta.go). The document itself is downloaded at
# generation time and is NOT embedded or committed.
API_VERSION ?= v1.21.2

SWAGGER_JSON = .cache/swagger-$(API_VERSION).json
SWAGGER_URL = https://raw.githubusercontent.com/kubernetes/kubernetes/refs/tags/$(API_VERSION)/api/openapi-spec/swagger.json

.PHONY: all
all: \
	kustomizationapi/swagger.go \
	zz_generated_patchmeta.go

.PHONY: clean
clean:
	rm -f kustomizationapi/swagger.go
	rm -rf .cache

$(SWAGGER_JSON):
	mkdir -p .cache
	curl -sSfL $(SWAGGER_URL) -o $@

# Regenerate the precomputed patch-metadata table from the pinned
# kubernetes/kubernetes tag.
.PHONY: zz_generated_patchmeta.go
zz_generated_patchmeta.go: $(SWAGGER_JSON)
	go run ./internal/patchmetagen -k8s-version $(API_VERSION) $(SWAGGER_JSON) > $@

# Verify the committed table is equivalent to the source schema document at
# the pinned tag (runs TestPrecomputedPatchMetaEquivalence, which skips when
# the document is not available).
.PHONY: verify-patchmeta
verify-patchmeta: $(SWAGGER_JSON)
	PATCHMETA_SWAGGER=$(CURDIR)/$(SWAGGER_JSON) go test . -run TestPrecomputedPatchMetaEquivalence -count=1 -v

$(MYGOBIN)/go-bindata:
	go install github.com/go-bindata/go-bindata/v3/go-bindata@latest

kustomizationapi/swagger.go: $(MYGOBIN)/go-bindata kustomizationapi/swagger.json
	$(MYGOBIN)/go-bindata \
		--pkg kustomizationapi \
		-o kustomizationapi/swagger.go \
		kustomizationapi/swagger.json
