// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"path/filepath"
	"strings"
	"testing"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func BenchmarkProtoUnmarshal(t *testing.B) {
	version := kubernetesOpenAPIDefaultVersion

	// parse the swagger, this should never fail
	assetName := filepath.Join(
		"kubernetesapi",
		strings.ReplaceAll(version, ".", "_"),
		"swagger.pb")

	b := kubernetesapi.OpenAPIMustAsset[version](assetName)

	for i := 0; i < t.N; i++ {
		// We parse protobuf and get an openapiv2.Document here.
		if err := proto.Unmarshal(b, &openapi_v2.Document{}); err != nil {
			t.Fatalf("proto.Unmarshal failed: %v", err)
		}
	}
}

func BenchmarkPrecomputedIsNamespaceScoped(b *testing.B) {
	testcases := map[string]yaml.TypeMeta{
		"namespace scoped": {APIVersion: "apps/v1", Kind: "ControllerRevision"},
		"cluster scoped":   {APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
		"unknown resource": {APIVersion: "custom.io/v1", Kind: "Custom"},
	}
	for name, testcase := range testcases {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ResetOpenAPI()
				_, _ = IsNamespaceScoped(testcase)
			}
		})
	}
}
