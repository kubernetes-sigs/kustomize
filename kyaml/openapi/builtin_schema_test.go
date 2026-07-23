// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"bytes"
	"compress/gzip"
	"testing"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi/v1_21_2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestBuiltinOpenAPIBundle(t *testing.T) {
	bundle, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle)
	require.NoError(t, err)
	require.Equal(t, builtinopenapi.FormatVersion, bundle.FormatVersion)
	require.Equal(t, builtinopenapi.Coverage{Floor: "v1.21.2", Ceiling: "v1.36.2"}, bundle.Coverage)
	require.Len(t, bundle.Sources, 16)
	require.Len(t, bundle.Definitions, 1226)
	require.Len(t, bundle.Resources, 484)

	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)
	require.NoError(t, parseBuiltinBundle(builtinKubernetesOpenAPIBundle))
	require.Len(t, globalSchema.schema.Definitions, len(bundle.Definitions))

	definitions := 0
	scopes := 0
	for _, resource := range bundle.Resources {
		typeMeta := yaml.TypeMeta{APIVersion: resource.APIVersion, Kind: resource.Kind}
		if resource.Definition != "" {
			definitions++
			require.NotNil(t, globalSchema.schemaByResourceType[typeMeta], "%v", typeMeta)
		}
		if resource.Scope != builtinopenapi.ScopeUnknown {
			scopes++
			_, found := globalSchema.namespaceabilityByResourceType[typeMeta]
			require.True(t, found, "%v", typeMeta)
		}
	}
	require.Len(t, globalSchema.schemaByResourceType, definitions)
	require.Len(t, globalSchema.namespaceabilityByResourceType, scopes)
}

func TestBuiltinOpenAPIBundlePreservesLegacyGVKs(t *testing.T) {
	document := &openapi_v2.Document{}
	require.NoError(t, proto.Unmarshal(v1_21_2.MustAsset(
		"kubernetesapi/v1_21_2/swagger.pb"), document))
	var swagger spec.Swagger
	ok, err := swagger.FromGnostic(document)
	require.NoError(t, err)
	require.True(t, ok)

	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)
	AddDefinitions(swagger.Definitions)
	findNamespaceability(swagger.Paths)
	legacySchemas := make(map[yaml.TypeMeta]spec.Schema, len(globalSchema.schemaByResourceType))
	for typeMeta, schema := range globalSchema.schemaByResourceType {
		legacySchemas[typeMeta] = *schema
	}
	legacyScopes := make(map[yaml.TypeMeta]bool, len(globalSchema.namespaceabilityByResourceType))
	for typeMeta, namespaced := range globalSchema.namespaceabilityByResourceType {
		legacyScopes[typeMeta] = namespaced
	}

	ResetOpenAPI()
	require.NoError(t, parseBuiltinBundle(builtinKubernetesOpenAPIBundle))
	for typeMeta := range legacySchemas {
		require.Contains(t, globalSchema.schemaByResourceType, typeMeta)
	}
	for typeMeta, namespaced := range legacyScopes {
		require.Equal(t, namespaced, globalSchema.namespaceabilityByResourceType[typeMeta], "%v", typeMeta)
	}
}

func TestBuiltinOpenAPIBundleContainsOldIntermediateAndCurrentGVKs(t *testing.T) {
	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)

	legacy := SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "extensions/v1beta1",
		Kind:       "Ingress",
	})
	require.NotNil(t, legacy)
	intermediate := SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "flowcontrol.apiserver.k8s.io/v1beta2",
		Kind:       "FlowSchema",
	})
	require.NotNil(t, intermediate)
	current := SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "resource.k8s.io/v1",
		Kind:       "ResourceClaim",
	})
	require.NotNil(t, current)
}

func TestBuiltinOpenAPIVersionAliases(t *testing.T) {
	require.Equal(t, "v1.36", DefaultOpenAPI)
	require.Equal(t, "{title:Kubernetes,version:v1.36}", BuiltinSchemaInfo)
	require.True(t, hasBuiltinOpenAPIVersion(DefaultOpenAPI))
	require.True(t, hasBuiltinOpenAPIVersion(legacyOpenAPI))
	require.False(t, hasBuiltinOpenAPIVersion("v1.35"))
}

func TestLegacyBuiltinOpenAPIVersionLoadsUnion(t *testing.T) {
	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)
	require.NoError(t, SetSchema(map[string]string{"version": legacyOpenAPI}, nil, false))
	require.Equal(t, legacyOpenAPI, GetSchemaVersion())
	require.NotNil(t, SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "admissionregistration.k8s.io/v1",
		Kind:       "MutatingAdmissionPolicy",
	}))
}

func TestDecodeBuiltinOpenAPIBundleRejectsInvalidData(t *testing.T) {
	_, err := decodeBuiltinBundle([]byte("not gzip"))
	require.Error(t, err)

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err = writer.Write([]byte(`{"formatVersion":2}`))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	_, err = decodeBuiltinBundle(compressed.Bytes())
	require.ErrorContains(t, err, "unsupported built-in OpenAPI bundle format")
}

func TestBuiltinKustomizationSchema(t *testing.T) {
	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)
	schema := SchemaForResourceType(yaml.TypeMeta{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
	})
	require.NotNil(t, schema)
	strategy, key := schema.Field("configMapGenerator").PatchStrategyAndKey()
	require.Equal(t, "merge", strategy)
	require.Equal(t, "name", key)
	strategy, key = schema.Field("secretGenerator").PatchStrategyAndKey()
	require.Equal(t, "merge", strategy)
	require.Equal(t, "name", key)
}
