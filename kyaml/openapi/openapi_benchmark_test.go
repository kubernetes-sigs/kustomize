// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"path/filepath"
	"testing"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
)

func BenchmarkProtoUnmarshal(t *testing.B) {
	version := kubernetesOpenAPIDefaultVersion

	// parse the swagger, this should never fail
	assetName := filepath.Join(
		"kubernetesapi",
		version,
		"swagger.pb")

	b := kubernetesapi.OpenAPIMustAsset[version](assetName)

	for i := 0; i < t.N; i++ {
		// We parse protobuf and get an openapiv2.Document here.
		if err := proto.Unmarshal(b, &openapi_v2.Document{}); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
	}
}
