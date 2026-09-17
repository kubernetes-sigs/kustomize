# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

.PHONY: all
all: generate

.PHONY: generate
generate:
	GOWORK=off go generate .

.PHONY: verify-generated
verify-generated: generate
	git diff --exit-code -- \
		builtin_schema_provenance_test.go \
		builtin_schema_scope.go \
		builtin_schema_version.go \
		kubernetesapi/data/kubernetes-openapi-union.bundle-v1.json.gz

.PHONY: verify
verify:
	GOWORK=off go test ./...
