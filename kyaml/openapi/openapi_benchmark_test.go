// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/proto"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi/v1218pb"
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

func BenchmarkOpenAPIV2ParseDocument(t *testing.B) {
	version := kubernetesOpenAPIDefaultVersion

	assetName := filepath.Join(
		"kubernetesapi",
		version,
		"swagger.json")

	b := kubernetesapi.OpenAPIMustAsset[version](assetName)

	for i := 0; i < t.N; i++ {
		// We parse JSON and get an openapiv2.Document here.
		if _, err := openapi_v2.ParseDocument(b); err != nil {
			t.Fatalf("openapi_v2.ParseDocument failed: %v", err)
		}
	}
}

func BenchmarkProtoUnmarshal(t *testing.B) {
	assetName := filepath.Join(
		"kubernetesapi",
		"v1218pb",
		"swagger.pb")

	b := v1218pb.MustAsset(assetName)

	for i := 0; i < t.N; i++ {
		// We parse protobuf and get an openapiv2.Document here.
		if err := proto.Unmarshal(b, &openapi_v2.Document{}); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
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
