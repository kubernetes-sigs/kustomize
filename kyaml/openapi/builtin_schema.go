// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import _ "embed"

//go:generate go run ./cmd/openapi-bundle -manifest kubernetesapi/sources/builtin-union-v1.36.json -output kubernetesapi/data/kubernetes-openapi-union-v1.36.bundle-v1.json.gz -scope-output builtin_schema_scope.go

const (
	// DefaultOpenAPI identifies the built-in union schema. Its suffix is the
	// newest Kubernetes minor represented by the artifact; exact source
	// coverage is recorded in the bundle metadata.
	DefaultOpenAPI = "v1.36"

	// legacyOpenAPI is retained as an accepted alias for configurations that
	// selected the previously embedded schema explicitly.
	legacyOpenAPI = "v1.21.2"

	// BuiltinSchemaInfo is the value printed by `kustomize openapi info`.
	BuiltinSchemaInfo = "{title:Kubernetes,version:" + DefaultOpenAPI + "}"
)

//go:embed kubernetesapi/data/kubernetes-openapi-union-v1.36.bundle-v1.json.gz
var builtinKubernetesOpenAPIBundle []byte

//go:embed kustomizationapi/swagger.json
var builtinKustomizationOpenAPI []byte

func hasBuiltinOpenAPIVersion(version string) bool {
	return version == DefaultOpenAPI || version == legacyOpenAPI
}
