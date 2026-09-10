// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
	"sigs.k8s.io/kustomize/kyaml/openapi/kubernetesapi/v1_21_2"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type builtinProvenanceAPI struct {
	apiVersion string
	kind       string
	definition string
}

type builtinSourceProvenance struct {
	kubernetesVersion string
	gitCommit         string
	sha256            string
	apis              []builtinProvenanceAPI
}

func TestBuiltinOpenAPIBundle(t *testing.T) {
	bundle, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle)
	require.NoError(t, err)
	require.Equal(t, builtinopenapi.FormatVersion, bundle.FormatVersion)
	require.Equal(t, builtinopenapi.Coverage{
		Floor:   generatedBuiltinSourceProvenance[0].kubernetesVersion,
		Ceiling: generatedBuiltinSourceProvenance[len(generatedBuiltinSourceProvenance)-1].kubernetesVersion,
	}, bundle.Coverage)
	require.Len(t, bundle.Sources, len(generatedBuiltinSourceProvenance))
	require.NotEmpty(t, bundle.Definitions)
	require.NotEmpty(t, bundle.Resources)

	definitionResources := make(map[yaml.TypeMeta]string)
	scopes := 0
	for _, resource := range bundle.Resources {
		typeMeta := yaml.TypeMeta{APIVersion: resource.APIVersion, Kind: resource.Kind}
		if resource.Definition != "" {
			definitionResources[typeMeta] = resource.Definition
		}
		if resource.Scope != builtinopenapi.ScopeUnknown {
			scopes++
		}
	}

	// A bundle must carry the same GVK-to-definition mapping in both its
	// definitions and its resource inventory. The runtime indexes schemas
	// exclusively from the definitions' GVK extensions.
	ResetOpenAPI()
	t.Cleanup(ResetOpenAPI)
	AddDefinitions(bundle.Definitions)
	require.Len(t, globalSchema.schemaByResourceType, len(definitionResources))
	for typeMeta, definitionName := range definitionResources {
		indexed := globalSchema.schemaByResourceType[typeMeta]
		require.NotNil(t, indexed, "%v", typeMeta)
		require.Equal(t, bundle.Definitions[definitionName], *indexed, "%v", typeMeta)
	}

	ResetOpenAPI()
	require.NoError(t, parseBuiltinBundle(builtinKubernetesOpenAPIBundle))
	require.Len(t, globalSchema.schema.Definitions, len(bundle.Definitions))
	require.Len(t, globalSchema.schemaByResourceType, len(definitionResources))
	require.Len(t, globalSchema.namespaceabilityByResourceType, scopes)
	for _, resource := range bundle.Resources {
		if resource.Scope == builtinopenapi.ScopeUnknown {
			continue
		}
		typeMeta := yaml.TypeMeta{APIVersion: resource.APIVersion, Kind: resource.Kind}
		_, found := globalSchema.namespaceabilityByResourceType[typeMeta]
		require.True(t, found, "%v", typeMeta)
	}
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
	bundle, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle)
	require.NoError(t, err)
	ceiling := bundle.Coverage.Ceiling
	defaultMinor := ceiling[:strings.LastIndex(ceiling, ".")]
	require.Equal(t, defaultMinor, DefaultOpenAPI)
	require.Equal(t, "{title:Kubernetes,version:"+DefaultOpenAPI+"}", BuiltinSchemaInfo)
	require.True(t, hasBuiltinOpenAPIVersion(DefaultOpenAPI))
	require.True(t, hasBuiltinOpenAPIVersion("builtin"))
	require.True(t, hasBuiltinOpenAPIVersion("v1.21.2"))
	require.True(t, hasBuiltinOpenAPIVersion("v1.35"))
	require.False(t, hasBuiltinOpenAPIVersion("v2.1"))
}

func TestBuiltinOpenAPIVersionAliasesLoadUnion(t *testing.T) {
	for _, alias := range []string{"builtin", "v1.21.2"} {
		t.Run(alias, func(t *testing.T) {
			ResetOpenAPI()
			t.Cleanup(ResetOpenAPI)
			require.NoError(t, SetSchema(map[string]string{"version": alias}, nil, false))
			require.Equal(t, alias, GetSchemaVersion())
			require.NotNil(t, SchemaForResourceType(yaml.TypeMeta{
				APIVersion: "admissionregistration.k8s.io/v1",
				Kind:       "MutatingAdmissionPolicy",
			}))
		})
	}
}

func TestBuiltinOpenAPISourceProvenance(t *testing.T) {
	bundle, err := decodeBuiltinBundle(builtinKubernetesOpenAPIBundle)
	require.NoError(t, err)
	bundleResources := make(map[yaml.TypeMeta]builtinopenapi.Resource, len(bundle.Resources))
	for _, resource := range bundle.Resources {
		bundleResources[yaml.TypeMeta{APIVersion: resource.APIVersion, Kind: resource.Kind}] = resource
	}
	require.Len(t, generatedBuiltinSourceProvenance, len(bundle.Sources))
	for i, provenance := range generatedBuiltinSourceProvenance {
		source := bundle.Sources[len(bundle.Sources)-1-i]
		require.Equal(t, source.KubernetesVersion, provenance.kubernetesVersion)
		require.Equal(t, source.GitCommit, provenance.gitCommit)
		require.Equal(t, source.SHA256, provenance.sha256)
		require.NotEmpty(t, provenance.apis, "Kubernetes %s", provenance.kubernetesVersion)
		require.LessOrEqual(t, len(provenance.apis), 3)
		for _, api := range provenance.apis {
			resource, found := bundleResources[yaml.TypeMeta{APIVersion: api.apiVersion, Kind: api.kind}]
			require.True(t, found, "%s %s/%s", provenance.kubernetesVersion, api.apiVersion, api.kind)
			require.Equal(t, api.definition, resource.Definition)
			require.Contains(t, bundle.Definitions, api.definition)
		}
	}
}

func TestDecodeBuiltinOpenAPIBundleRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name              string
		data              []byte
		alreadyCompressed bool
		corruptChecksum   bool
		errorContains     string
	}{
		{
			name:              "invalid gzip",
			data:              []byte("not gzip"),
			alreadyCompressed: true,
		},
		{
			name: "invalid JSON",
			data: []byte(`{"formatVersion":`),
		},
		{
			name:          "multiple JSON values",
			data:          []byte(`{} {}`),
			errorContains: "multiple JSON values",
		},
		{
			name: "invalid trailing data",
			data: []byte(`{} trailing`),
		},
		{
			name:              "invalid checksum",
			data:              builtinKubernetesOpenAPIBundle,
			alreadyCompressed: true,
			corruptChecksum:   true,
			errorContains:     "gzip: invalid checksum",
		},
		{
			name:          "invalid bundle",
			data:          []byte(`{"formatVersion":2}`),
			errorContains: "unsupported built-in OpenAPI bundle format",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compressed := bytes.Clone(test.data)
			if !test.alreadyCompressed {
				compressed = gzipBytes(t, test.data)
			}
			if test.corruptChecksum {
				compressed[len(compressed)-8] ^= 0xff
			}
			_, err := decodeBuiltinBundle(compressed)
			if test.errorContains == "" {
				require.Error(t, err)
			} else {
				require.ErrorContains(t, err, test.errorContains)
			}
		})
	}
}

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return compressed.Bytes()
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
