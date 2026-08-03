// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func BenchmarkBuiltinBundleDecode(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle); err != nil {
			b.Fatalf("decodeBuiltinBundle failed: %v", err)
		}
	}
}

func BenchmarkFirstSchemaLookup(b *testing.B) {
	b.ReportAllocs()
	typeMeta := yaml.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"}
	for i := 0; i < b.N; i++ {
		ResetOpenAPI()
		if schema := SchemaForResourceType(typeMeta); schema == nil {
			b.Fatal("Deployment schema was not found")
		}
	}
	ResetOpenAPI()
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
