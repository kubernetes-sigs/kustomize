// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"path/filepath"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
)

// Benchmark for swagger parsing (UnmarshalJSON)
func BenchmarkSwaggerUnmarshalJSON(t *testing.B) {
	version := kubernetesOpenAPIDefaultVersion

	// parse the swagger, this should never fail
	assetName := filepath.Join(
		"kubernetesapi",
		version,
		"swagger.json")

	b := kubernetesapi.OpenAPIMustAsset[version](assetName)

	for i := 0; i < t.N; i++ {
		var swagger spec.Swagger
		if err := swagger.UnmarshalJSON(b); err != nil {
			t.Fatalf("swagger.UnmarshalJSON failed: %v", err)
		}
	}
}

// Benchmark for loading assets packed into the binary
func BenchmarkAssetRead(t *testing.B) {
	for i := 0; i < t.N; i++ {
		version := kubernetesOpenAPIDefaultVersion

		// parse the swagger, this should never fail
		assetName := filepath.Join(
			"kubernetesapi",
			version,
			"swagger.json")

		kubernetesapi.OpenAPIMustAsset[version](assetName)
	}
}
