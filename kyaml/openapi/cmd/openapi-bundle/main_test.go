// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
)

func TestGeneratedBundleIsCurrentAndDeterministic(t *testing.T) {
	manifest := filepath.Join("..", "..", "kubernetesapi", "sources", "builtin-union-v1.36.json")
	checkedIn := filepath.Join("..", "..", "kubernetesapi", "data",
		"kubernetes-openapi-union-v1.36.bundle-v1.json.gz")
	checkedInScopes := filepath.Join("..", "..", "builtin_schema_scope.go")
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "first.json.gz")
	second := filepath.Join(tempDir, "second.json.gz")
	firstScopes := filepath.Join(tempDir, "first.go")
	secondScopes := filepath.Join(tempDir, "second.go")

	for i, output := range []string{first, second} {
		scopeOutput := []string{firstScopes, secondScopes}[i]
		require.NoError(t, run(options{
			manifest:    manifest,
			output:      output,
			scopeOutput: scopeOutput,
		}))
	}

	want, err := os.ReadFile(checkedIn)
	require.NoError(t, err)
	got, err := os.ReadFile(first)
	require.NoError(t, err)
	again, err := os.ReadFile(second)
	require.NoError(t, err)
	require.Equal(t, want, got, "checked-in bundle is stale")
	require.Equal(t, got, again, "bundle generation is not deterministic")
	wantScopes, err := os.ReadFile(checkedInScopes)
	require.NoError(t, err)
	gotScopes, err := os.ReadFile(firstScopes)
	require.NoError(t, err)
	againScopes, err := os.ReadFile(secondScopes)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Equal(t, wantScopes, gotScopes, "checked-in resource-scope index is stale")
	}
	require.Equal(t, gotScopes, againScopes, "resource-scope generation is not deterministic")

	reader, err := gzip.NewReader(bytes.NewReader(got))
	require.NoError(t, err)
	require.True(t, reader.ModTime.IsZero())
	require.Empty(t, reader.Name)
	require.Empty(t, reader.Comment)
	decoder := json.NewDecoder(reader)
	var bundle builtinopenapi.Bundle
	require.NoError(t, decoder.Decode(&bundle))
	var trailing interface{}
	require.ErrorIs(t, decoder.Decode(&trailing), io.EOF)
	require.NoError(t, reader.Close())
	require.NoError(t, bundle.Validate())
	require.Equal(t, builtinopenapi.Coverage{Floor: "v1.21.2", Ceiling: "v1.36.2"}, bundle.Coverage)
	require.Len(t, bundle.Sources, 16)
	require.Len(t, bundle.Definitions, 1226)
	require.Len(t, bundle.Resources, 484)
	require.Equal(t, "dcede2063da1d7ad62ecb5af8adb6d7fabd0b52385a7fa0048afb491dac90450",
		bundle.Sources[0].SHA256)
	require.Equal(t, "128a984dbb5a4e5ceceef9dea0db575267678d333f53ed606300a2132d2539cc",
		bundle.Sources[len(bundle.Sources)-1].SHA256)
	requireManifestInventoryIncluded(t, manifest, &bundle)
}

func requireManifestInventoryIncluded(t *testing.T, manifestPath string, bundle *builtinopenapi.Bundle) {
	t.Helper()
	manifestBytes, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var manifest sourceManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))

	bundleResources := make(map[string]builtinopenapi.Resource, len(bundle.Resources))
	for _, resource := range bundle.Resources {
		bundleResources[resourceKey(resource.APIVersion, resource.Kind)] = resource
	}
	manifestDir := filepath.Dir(manifestPath)
	for _, source := range manifest.Sources {
		input, err := readInput(filepath.Join(manifestDir, source.File))
		require.NoError(t, err)
		swagger, err := decodeSwagger(input)
		require.NoError(t, err)
		for definitionName := range swagger.Definitions {
			require.Contains(t, bundle.Definitions, definitionName,
				"definition from Kubernetes %s", source.KubernetesVersion)
		}
		resources, err := collectResources(swagger)
		require.NoError(t, err)
		for _, resource := range resources {
			actual, found := bundleResources[resourceKey(resource.APIVersion, resource.Kind)]
			require.True(t, found, "GVK %s/%s from Kubernetes %s",
				resource.APIVersion, resource.Kind, source.KubernetesVersion)
			if resource.Scope != builtinopenapi.ScopeUnknown {
				require.Equal(t, resource.Scope, actual.Scope, "GVK %s/%s from Kubernetes %s",
					resource.APIVersion, resource.Kind, source.KubernetesVersion)
			}
		}
	}
}

func TestLegacyProtoOutputIsDeterministic(t *testing.T) {
	source := filepath.Join("..", "..", "kubernetesapi", "v1_21_2", "swagger.pb.gz")
	legacy := filepath.Join(t.TempDir(), "swagger.pb.gz")
	require.NoError(t, run(options{
		input:             source,
		output:            filepath.Join(t.TempDir(), "single.json.gz"),
		legacyProtoOutput: legacy,
		kubernetesVersion: "v1.21.2",
	}))

	sourceArchive, err := os.ReadFile(source)
	require.NoError(t, err)
	legacyArchive, err := os.ReadFile(legacy)
	require.NoError(t, err)
	require.Equal(t, sourceArchive, legacyArchive)
}

func TestCompileSourcesUsesWholeNewestDefinitionAndUnionsGVKs(t *testing.T) {
	newest := sourceDocument{
		kubernetesVersion: "v1.36.2",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.36.2"},
  "paths": {},
  "definitions": {
    "shared": {"type": "object", "properties": {"newest": {"type": "string"}}},
    "current": {
      "type": "object",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1", "kind": "Current"}],
      "properties": {"spec": {"$ref": "#/definitions/shared"}}
    }
  }
}`),
	}
	oldest := sourceDocument{
		kubernetesVersion: "v1.21.2",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.21.2"},
  "paths": {},
  "definitions": {
    "shared": {"type": "object", "properties": {"oldest": {"type": "integer"}}},
    "legacy": {
      "type": "object",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1beta1", "kind": "Legacy"}],
      "properties": {"spec": {"$ref": "#/definitions/shared"}}
    }
  }
}`),
	}

	bundle, err := compileSources([]sourceDocument{newest, oldest}, builtinopenapi.Coverage{
		Floor: "v1.21.2", Ceiling: "v1.36.2",
	})
	require.NoError(t, err)
	require.Len(t, bundle.Resources, 2)
	shared := bundle.Definitions["shared"]
	require.Contains(t, shared.Properties, "newest")
	require.NotContains(t, shared.Properties, "oldest", "definitions must not be merged field-by-field")
}

func TestCompileSourcesValidatesEachSourceReferences(t *testing.T) {
	broken := sourceDocument{
		kubernetesVersion: "v1.36.2",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.36.2"},
  "paths": {},
  "definitions": {"root": {"$ref": "#/definitions/only-in-older-source"}}
}`),
	}
	older := sourceDocument{
		kubernetesVersion: "v1.21.2",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.21.2"},
  "paths": {},
  "definitions": {"only-in-older-source": {"type": "object"}}
}`),
	}

	_, err := compileSources([]sourceDocument{broken, older}, builtinopenapi.Coverage{
		Floor: "v1.21.2", Ceiling: "v1.36.2",
	})
	require.ErrorContains(t, err, "validate Kubernetes v1.36.2 definitions")
	require.ErrorContains(t, err, "only-in-older-source")
}

func TestCompileSourcesRejectsScopeConflict(t *testing.T) {
	source := func(version, path string) sourceDocument {
		return sourceDocument{
			kubernetesVersion: version,
			input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "` + version + `"},
  "paths": {
    "` + path + `": {
      "get": {
        "x-kubernetes-group-version-kind": {"group": "example.io", "version": "v1", "kind": "Widget"},
        "responses": {"200": {"description": "ok"}}
      }
    }
  },
  "definitions": {
    "widget": {
      "type": "object",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1", "kind": "Widget"}]
    }
  }
}`),
		}
	}
	newest := source("v1.36.2", "/apis/example.io/v1/namespaces/{namespace}/widgets")
	oldest := source("v1.21.2", "/apis/example.io/v1/widgets")

	_, err := compileSources([]sourceDocument{newest, oldest}, builtinopenapi.Coverage{
		Floor: "v1.21.2", Ceiling: "v1.36.2",
	})
	require.ErrorContains(t, err, `has scopes "Namespaced" and "Cluster"`)
}

func TestValidateManifest(t *testing.T) {
	validSource := func(version string) sourceManifestEntry {
		return sourceManifestEntry{
			KubernetesVersion: version,
			GitCommit:         strings.Repeat("a", 40),
			File:              version + ".json.gz",
			SHA256:            strings.Repeat("b", 64),
		}
	}
	tests := []struct {
		name    string
		sources []sourceManifestEntry
		wantErr string
	}{
		{name: "valid", sources: []sourceManifestEntry{validSource("v1.36.2"), validSource("v1.35.6")}},
		{name: "out of order", sources: []sourceManifestEntry{validSource("v1.35.6"), validSource("v1.36.2")}, wantErr: "newest-first"},
		{name: "duplicate", sources: []sourceManifestEntry{validSource("v1.36.2"), validSource("v1.36.2")}, wantErr: "newest-first"},
		{name: "duplicate minor", sources: []sourceManifestEntry{validSource("v1.36.2"), validSource("v1.36.1")}, wantErr: "exactly one snapshot"},
		{name: "minor gap", sources: []sourceManifestEntry{validSource("v1.36.2"), validSource("v1.34.9")}, wantErr: "exactly one snapshot"},
		{name: "invalid version", sources: []sourceManifestEntry{validSource("1.36")}, wantErr: "invalid Kubernetes version"},
		{name: "uppercase digest", sources: []sourceManifestEntry{{
			KubernetesVersion: "v1.36.2",
			GitCommit:         strings.Repeat("A", 40),
			File:              "v1.36.2.json.gz",
			SHA256:            strings.Repeat("B", 64),
		}}, wantErr: "lowercase"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manifest := sourceManifest{
				Repository:  "https://github.com/kubernetes/kubernetes",
				OpenAPIPath: "api/openapi-spec/swagger.json",
				Sources:     tc.sources,
			}
			err := validateManifest(&manifest)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestWriteGzipUsesStableHeader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.gz")
	require.NoError(t, writeGzip(path, []byte("data")))
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	reader, err := gzip.NewReader(bytes.NewReader(b))
	require.NoError(t, err)
	require.Equal(t, time.Time{}, reader.ModTime)
	require.Empty(t, reader.Name)
	require.Empty(t, reader.Comment)
	require.Equal(t, byte(255), reader.OS)
	require.NoError(t, reader.Close())
}
