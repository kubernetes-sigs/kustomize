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
	require.Equal(t, builtinopenapi.Coverage{Floor: "v1.21.2", Ceiling: "v1.21.2"}, bundle.Coverage)
	require.Len(t, bundle.Definitions, 618)
	require.Len(t, bundle.Resources, 275)

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

func TestBuiltinOpenAPIBundleMatchesLegacySchema(t *testing.T) {
	document := &openapi_v2.Document{}
	require.NoError(t, proto.Unmarshal(v1_21_2.MustAsset(
		"kubernetesapi/v1_21_2/swagger.pb"), document))
	var swagger spec.Swagger
	ok, err := swagger.FromGnostic(document)
	require.NoError(t, err)
	require.True(t, ok)

	bundle, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle)
	require.NoError(t, err)
	require.Equal(t, swagger.Definitions, bundle.Definitions)

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
	require.Len(t, globalSchema.schemaByResourceType, len(legacySchemas))
	for typeMeta, schema := range legacySchemas {
		require.Equal(t, schema, *globalSchema.schemaByResourceType[typeMeta], "%v", typeMeta)
	}
	require.Equal(t, legacyScopes, globalSchema.namespaceabilityByResourceType)
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
