// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import _ "embed"

//go:generate go run ./cmd/openapi-bundle -manifest kubernetesapi/builtin-versions.json -lock kubernetesapi/data/kubernetes-openapi-union.bundle-v1.json.gz -output kubernetesapi/data/kubernetes-openapi-union.bundle-v1.json.gz -scope-output builtin_schema_scope.go -version-output builtin_schema_version.go -provenance-output builtin_schema_provenance_test.go

//go:embed kubernetesapi/data/kubernetes-openapi-union.bundle-v1.json.gz
var builtinKubernetesOpenAPIBundle []byte

//go:embed kustomizationapi/swagger.json
var builtinKustomizationOpenAPI []byte
