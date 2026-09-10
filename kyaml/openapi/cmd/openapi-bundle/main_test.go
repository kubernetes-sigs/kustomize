// Copyright 2026 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/openapi/internal/builtinopenapi"
)

type fakeSourceProvider struct {
	commits         map[string]string
	inputs          map[string][]byte
	resolved        []string
	fetchedExpected map[string]string
}

func (p *fakeSourceProvider) resolve(_ context.Context, version string) (string, error) {
	p.resolved = append(p.resolved, version)
	return p.commits[version], nil
}

func (p *fakeSourceProvider) fetch(
	_ context.Context, version string, commit string, expectedSHA256 string,
) ([]byte, error) {
	if p.fetchedExpected == nil {
		p.fetchedExpected = make(map[string]string)
	}
	p.fetchedExpected[version] = expectedSHA256
	if expectedCommit := p.commits[version]; expectedCommit != commit {
		return nil, errors.New("unexpected commit")
	}
	return p.inputs[version], nil
}

func TestGeneratedBundleIsCurrentAndDeterministic(t *testing.T) {
	manifest := filepath.Join("..", "..", "kubernetesapi", "builtin-versions.json")
	checkedIn := filepath.Join("..", "..", "kubernetesapi", "data",
		"kubernetes-openapi-union.bundle-v1.json.gz")
	checkedInScopes := filepath.Join("..", "..", "builtin_schema_scope.go")
	checkedInVersions := filepath.Join("..", "..", "builtin_schema_version.go")
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "first.json.gz")
	second := filepath.Join(tempDir, "second.json.gz")
	firstScopes := filepath.Join(tempDir, "first.go")
	secondScopes := filepath.Join(tempDir, "second.go")
	generatedVersions := filepath.Join(tempDir, "versions.go")
	generatedVersionsAgain := filepath.Join(tempDir, "versions-again.go")

	manifestBytes, err := os.ReadFile(manifest)
	require.NoError(t, err)
	var sourceList versionManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &sourceList))
	require.NoError(t, validateManifest(&sourceList))
	checkedInBundle, err := readInput(checkedIn)
	require.NoError(t, err)
	var lockedBundle builtinopenapi.Bundle
	require.NoError(t, json.Unmarshal(checkedInBundle, &lockedBundle))
	require.NoError(t, lockedBundle.Validate())

	for i, output := range []string{first, second} {
		scopeOutput := []string{firstScopes, secondScopes}[i]
		require.NoError(t, writeBundle(output, &lockedBundle))
		require.NoError(t, writeScopeFile(scopeOutput, &lockedBundle))
	}
	require.NoError(t, writeVersionFile(generatedVersions, &sourceList))
	require.NoError(t, writeVersionFile(generatedVersionsAgain, &sourceList))

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
	wantVersions, err := os.ReadFile(checkedInVersions)
	require.NoError(t, err)
	gotVersions, err := os.ReadFile(generatedVersions)
	require.NoError(t, err)
	againVersions, err := os.ReadFile(generatedVersionsAgain)
	require.NoError(t, err)
	if runtime.GOOS != "windows" {
		require.Equal(t, wantVersions, gotVersions, "checked-in built-in version metadata is stale")
	}
	require.Equal(t, gotVersions, againVersions, "built-in version metadata generation is not deterministic")

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
	require.Equal(t, builtinopenapi.Coverage{
		Floor: sourceList.Versions[0], Ceiling: sourceList.Versions[len(sourceList.Versions)-1],
	}, bundle.Coverage)
	require.Len(t, bundle.Sources, len(sourceList.Versions))
	require.NotEmpty(t, bundle.Definitions)
	require.NotEmpty(t, bundle.Resources)
	for _, source := range bundle.Sources {
		require.Len(t, source.GitCommit, 40)
	}
	requireManifestInventoryIncluded(t, manifest, &bundle)
}

func requireManifestInventoryIncluded(t *testing.T, manifestPath string, bundle *builtinopenapi.Bundle) {
	t.Helper()
	manifestBytes, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var manifest versionManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))

	actualVersions := make([]string, len(bundle.Sources))
	for i, source := range bundle.Sources {
		actualVersions[len(bundle.Sources)-1-i] = source.KubernetesVersion
	}
	require.Equal(t, manifest.Versions, actualVersions)
}

func TestCompileManifestUsesExistingBundleAsSourceLock(t *testing.T) {
	oldCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	oldInput := []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.35.0"},
  "paths": {},
  "definitions": {
    "old": {
      "type": "object",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1", "kind": "Old"}]
    }
  }
}`)
	newInput := []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.36.0"},
  "paths": {},
  "definitions": {
    "new": {
      "type": "object",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1", "kind": "New"}]
    }
  }
}`)
	lockedBundle, err := compileSources([]sourceDocument{{
		kubernetesVersion: "v1.35.0",
		gitCommit:         oldCommit,
		input:             oldInput,
	}}, builtinopenapi.Coverage{Floor: "v1.35.0", Ceiling: "v1.35.0"})
	require.NoError(t, err)
	testDir := t.TempDir()
	bundlePath := filepath.Join(testDir, "bundle.json.gz")
	require.NoError(t, writeBundle(bundlePath, lockedBundle))
	manifestPath := filepath.Join(testDir, "versions.json")
	manifestBytes, err := json.Marshal(versionManifest{Versions: []string{"v1.35.0", "v1.36.0"}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, manifestBytes, 0o600))
	provider := &fakeSourceProvider{
		commits: map[string]string{"v1.35.0": oldCommit, "v1.36.0": newCommit},
		inputs:  map[string][]byte{"v1.35.0": oldInput, "v1.36.0": newInput},
	}

	manifest, bundle, provenance, err := compileManifest(manifestPath, bundlePath, provider)
	require.NoError(t, err)
	require.Equal(t, []string{"v1.35.0", "v1.36.0"}, manifest.Versions)
	require.Equal(t, []string{"v1.36.0"}, provider.resolved)
	require.NotEmpty(t, provider.fetchedExpected["v1.35.0"])
	require.Empty(t, provider.fetchedExpected["v1.36.0"])
	require.Equal(t, newCommit, bundle.Sources[0].GitCommit)
	require.Equal(t, oldCommit, bundle.Sources[1].GitCommit)
	require.Len(t, provenance, 2)
	require.NotEmpty(t, provenance[0].apis)
	require.NotEmpty(t, provenance[1].apis)
}

func TestCompileManifestRequiresExistingBundleLock(t *testing.T) {
	testDir := t.TempDir()
	manifestPath := filepath.Join(testDir, "versions.json")
	manifestBytes, err := json.Marshal(versionManifest{Versions: []string{"v1.36.0"}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(manifestPath, manifestBytes, 0o600))
	manifest, bundle, provenance, err := compileManifest(
		manifestPath, filepath.Join(testDir, "missing.json.gz"), &fakeSourceProvider{})
	require.Nil(t, manifest)
	require.Nil(t, bundle)
	require.Nil(t, provenance)
	require.ErrorContains(t, err, "restore the generated bundle")
}

func TestCollectSourceProvenanceSelectsFirstSeenAPIs(t *testing.T) {
	metaDefinition := "io.k8s.apimachinery.pkg.apis.meta.v1.DeleteOptions"
	inventories := []sourceProvenance{
		{
			kubernetesVersion: "v1.22.0",
			apis: []builtinopenapi.Resource{
				{APIVersion: "example.io/v1", Kind: "DeleteOptions", Definition: metaDefinition},
				{APIVersion: "example.io/v1", Kind: "New", Definition: "new"},
				{APIVersion: "example.io/v1", Kind: "NewList", Definition: "newList"},
				{APIVersion: "example.io/v1", Kind: "Shared", Definition: "shared"},
			},
		},
		{
			kubernetesVersion: "v1.21.0",
			apis: []builtinopenapi.Resource{
				{APIVersion: "example.io/v1beta1", Kind: "Old", Definition: "old"},
				{APIVersion: "example.io/v1", Kind: "Shared", Definition: "shared"},
			},
		},
	}

	actual, err := collectSourceProvenance(inventories)
	require.NoError(t, err)
	require.Equal(t, []sourceProvenance{
		{
			kubernetesVersion: "v1.21.0",
			apis: []builtinopenapi.Resource{
				{APIVersion: "example.io/v1beta1", Kind: "Old", Definition: "old"},
				{APIVersion: "example.io/v1", Kind: "Shared", Definition: "shared"},
			},
		},
		{
			kubernetesVersion: "v1.22.0",
			apis: []builtinopenapi.Resource{
				{APIVersion: "example.io/v1", Kind: "New", Definition: "new"},
			},
		},
	}, actual)
}

func TestCollectSourceProvenanceRequiresCuratedBaselineAPIs(t *testing.T) {
	provenance, err := collectSourceProvenance([]sourceProvenance{{
		kubernetesVersion: legacyBuiltinOpenAPIAlias,
		curated:           true,
		apis: []builtinopenapi.Resource{
			{APIVersion: "batch/v1", Kind: "CronJob", Definition: "wrong"},
		},
	}})
	require.Nil(t, provenance)
	require.ErrorContains(t, err, "does not contain curated provenance API batch/v1/CronJob")
}

func TestValidateManifestAgainstLockedSources(t *testing.T) {
	locked := []builtinopenapi.Source{
		{KubernetesVersion: "v1.36.0"},
		{KubernetesVersion: "v1.35.0"},
	}
	tests := []struct {
		name     string
		manifest versionManifest
		wantErr  string
	}{
		{
			name:     "append a minor",
			manifest: versionManifest{Versions: []string{"v1.35.0", "v1.36.0", "v1.37.0"}},
		},
		{
			name:     "remove newest minor",
			manifest: versionManifest{Versions: []string{"v1.35.0"}},
			wantErr:  "must not remove",
		},
		{
			name:     "remove oldest minor",
			manifest: versionManifest{Versions: []string{"v1.36.0", "v1.37.0"}},
			wantErr:  "must retain locked Kubernetes minor v1.35",
		},
		{
			name:     "replace a pinned source without correction",
			manifest: versionManifest{Versions: []string{"v1.35.1", "v1.36.0"}},
			wantErr:  "without an openapi-correction exception",
		},
		{
			name: "replace a pinned source with correction",
			manifest: versionManifest{
				Versions: []string{"v1.35.1", "v1.36.0"},
				PatchVersionExceptions: map[string]patchVersionException{
					"v1.35.1": {
						Reason:              patchVersionExceptionOpenAPICorrection,
						UpstreamPullRequest: "https://github.com/kubernetes/kubernetes/pull/123",
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateManifestAgainstLockedSources(&tc.manifest, locked)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestWriteVersionFilePreservesLegacyAliasIfBaselineSourceChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "versions.go")
	require.NoError(t, writeVersionFile(path, &versionManifest{Versions: []string{"v1.21.3"}}))
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(contents), `case "`+legacyBuiltinOpenAPIAlias+`":`)
}

func TestBuiltinUnionContainsAPIsIntroducedByEachMinor(t *testing.T) {
	type introducedAPI struct {
		apiVersion string
		kind       string
		definition string
	}
	type minorAPIs struct {
		kubernetesVersion string
		apis              []introducedAPI
	}

	// Keep this curated regression set oldest-first. The generated provenance
	// test data covers later sources automatically when a new minor is added.
	expected := []minorAPIs{
		{kubernetesVersion: "v1.21.2", apis: []introducedAPI{
			{apiVersion: "batch/v1", kind: "CronJob", definition: "io.k8s.api.batch.v1.CronJob"},
			{apiVersion: "discovery.k8s.io/v1", kind: "EndpointSlice", definition: "io.k8s.api.discovery.v1.EndpointSlice"},
			{apiVersion: "policy/v1", kind: "PodDisruptionBudget", definition: "io.k8s.api.policy.v1.PodDisruptionBudget"},
		}},
		{kubernetesVersion: "v1.22.0", apis: []introducedAPI{
			{apiVersion: "policy/v1", kind: "Eviction", definition: "io.k8s.api.policy.v1.Eviction"},
		}},
		{kubernetesVersion: "v1.23.0", apis: []introducedAPI{
			{apiVersion: "autoscaling/v2", kind: "HorizontalPodAutoscaler", definition: "io.k8s.api.autoscaling.v2.HorizontalPodAutoscaler"},
			{apiVersion: "flowcontrol.apiserver.k8s.io/v1beta2", kind: "FlowSchema", definition: "io.k8s.api.flowcontrol.v1beta2.FlowSchema"},
			{apiVersion: "flowcontrol.apiserver.k8s.io/v1beta2", kind: "PriorityLevelConfiguration", definition: "io.k8s.api.flowcontrol.v1beta2.PriorityLevelConfiguration"},
		}},
		{kubernetesVersion: "v1.24.0", apis: []introducedAPI{
			{apiVersion: "storage.k8s.io/v1", kind: "CSIStorageCapacity", definition: "io.k8s.api.storage.v1.CSIStorageCapacity"},
		}},
		{kubernetesVersion: "v1.25.0", apis: []introducedAPI{
			{apiVersion: "networking.k8s.io/v1alpha1", kind: "ClusterCIDR", definition: "io.k8s.api.networking.v1alpha1.ClusterCIDR"},
		}},
		{kubernetesVersion: "v1.26.2", apis: []introducedAPI{
			{
				apiVersion: "admissionregistration.k8s.io/v1alpha1",
				kind:       "ValidatingAdmissionPolicy",
				definition: "io.k8s.api.admissionregistration.v1alpha1.ValidatingAdmissionPolicy",
			},
			{apiVersion: "authentication.k8s.io/v1alpha1", kind: "SelfSubjectReview", definition: "io.k8s.api.authentication.v1alpha1.SelfSubjectReview"},
			{apiVersion: "resource.k8s.io/v1alpha1", kind: "ResourceClaim", definition: "io.k8s.api.resource.v1alpha1.ResourceClaim"},
		}},
		{kubernetesVersion: "v1.27.0", apis: []introducedAPI{
			{apiVersion: "certificates.k8s.io/v1alpha1", kind: "ClusterTrustBundle", definition: "io.k8s.api.certificates.v1alpha1.ClusterTrustBundle"},
			{apiVersion: "networking.k8s.io/v1alpha1", kind: "IPAddress", definition: "io.k8s.api.networking.v1alpha1.IPAddress"},
			{apiVersion: "resource.k8s.io/v1alpha2", kind: "ResourceClaim", definition: "io.k8s.api.resource.v1alpha2.ResourceClaim"},
		}},
		{kubernetesVersion: "v1.28.0", apis: []introducedAPI{
			{
				apiVersion: "admissionregistration.k8s.io/v1beta1",
				kind:       "ValidatingAdmissionPolicy",
				definition: "io.k8s.api.admissionregistration.v1beta1.ValidatingAdmissionPolicy",
			},
			{
				apiVersion: "admissionregistration.k8s.io/v1beta1",
				kind:       "ValidatingAdmissionPolicyBinding",
				definition: "io.k8s.api.admissionregistration.v1beta1.ValidatingAdmissionPolicyBinding",
			},
			{apiVersion: "authentication.k8s.io/v1", kind: "SelfSubjectReview", definition: "io.k8s.api.authentication.v1.SelfSubjectReview"},
		}},
		{kubernetesVersion: "v1.29.0", apis: []introducedAPI{
			{apiVersion: "flowcontrol.apiserver.k8s.io/v1", kind: "FlowSchema", definition: "io.k8s.api.flowcontrol.v1.FlowSchema"},
			{apiVersion: "networking.k8s.io/v1alpha1", kind: "ServiceCIDR", definition: "io.k8s.api.networking.v1alpha1.ServiceCIDR"},
			{apiVersion: "storage.k8s.io/v1alpha1", kind: "VolumeAttributesClass", definition: "io.k8s.api.storage.v1alpha1.VolumeAttributesClass"},
		}},
		{kubernetesVersion: "v1.30.0", apis: []introducedAPI{
			{apiVersion: "resource.k8s.io/v1alpha2", kind: "ResourceClaimParameters", definition: "io.k8s.api.resource.v1alpha2.ResourceClaimParameters"},
			{apiVersion: "resource.k8s.io/v1alpha2", kind: "ResourceClassParameters", definition: "io.k8s.api.resource.v1alpha2.ResourceClassParameters"},
			{apiVersion: "resource.k8s.io/v1alpha2", kind: "ResourceSlice", definition: "io.k8s.api.resource.v1alpha2.ResourceSlice"},
		}},
		{kubernetesVersion: "v1.31.0", apis: []introducedAPI{
			{apiVersion: "coordination.k8s.io/v1alpha1", kind: "LeaseCandidate", definition: "io.k8s.api.coordination.v1alpha1.LeaseCandidate"},
			{apiVersion: "networking.k8s.io/v1beta1", kind: "IPAddress", definition: "io.k8s.api.networking.v1beta1.IPAddress"},
			{apiVersion: "resource.k8s.io/v1alpha3", kind: "PodSchedulingContext", definition: "io.k8s.api.resource.v1alpha3.PodSchedulingContext"},
		}},
		{kubernetesVersion: "v1.32.0", apis: []introducedAPI{
			{
				apiVersion: "admissionregistration.k8s.io/v1alpha1",
				kind:       "MutatingAdmissionPolicy",
				definition: "io.k8s.api.admissionregistration.v1alpha1.MutatingAdmissionPolicy",
			},
			{
				apiVersion: "admissionregistration.k8s.io/v1alpha1",
				kind:       "MutatingAdmissionPolicyBinding",
				definition: "io.k8s.api.admissionregistration.v1alpha1.MutatingAdmissionPolicyBinding",
			},
			{apiVersion: "coordination.k8s.io/v1alpha2", kind: "LeaseCandidate", definition: "io.k8s.api.coordination.v1alpha2.LeaseCandidate"},
		}},
		{kubernetesVersion: "v1.33.0", apis: []introducedAPI{
			{apiVersion: "certificates.k8s.io/v1beta1", kind: "ClusterTrustBundle", definition: "io.k8s.api.certificates.v1beta1.ClusterTrustBundle"},
			{apiVersion: "networking.k8s.io/v1", kind: "ServiceCIDR", definition: "io.k8s.api.networking.v1.ServiceCIDR"},
			{apiVersion: "resource.k8s.io/v1alpha3", kind: "DeviceTaintRule", definition: "io.k8s.api.resource.v1alpha3.DeviceTaintRule"},
		}},
		{kubernetesVersion: "v1.34.0", apis: []introducedAPI{
			{apiVersion: "admissionregistration.k8s.io/v1beta1", kind: "MutatingAdmissionPolicy", definition: "io.k8s.api.admissionregistration.v1beta1.MutatingAdmissionPolicy"},
			{apiVersion: "certificates.k8s.io/v1alpha1", kind: "PodCertificateRequest", definition: "io.k8s.api.certificates.v1alpha1.PodCertificateRequest"},
			{apiVersion: "storage.k8s.io/v1", kind: "VolumeAttributesClass", definition: "io.k8s.api.storage.v1.VolumeAttributesClass"},
		}},
		{kubernetesVersion: "v1.35.0", apis: []introducedAPI{
			{apiVersion: "certificates.k8s.io/v1beta1", kind: "PodCertificateRequest", definition: "io.k8s.api.certificates.v1beta1.PodCertificateRequest"},
			{apiVersion: "scheduling.k8s.io/v1alpha1", kind: "Workload", definition: "io.k8s.api.scheduling.v1alpha1.Workload"},
			{apiVersion: "storagemigration.k8s.io/v1beta1", kind: "StorageVersionMigration", definition: "io.k8s.api.storagemigration.v1beta1.StorageVersionMigration"},
		}},
		{kubernetesVersion: "v1.36.0", apis: []introducedAPI{
			{apiVersion: "admissionregistration.k8s.io/v1", kind: "MutatingAdmissionPolicy", definition: "io.k8s.api.admissionregistration.v1.MutatingAdmissionPolicy"},
			{apiVersion: "resource.k8s.io/v1beta2", kind: "DeviceTaintRule", definition: "io.k8s.api.resource.v1beta2.DeviceTaintRule"},
			{apiVersion: "scheduling.k8s.io/v1alpha2", kind: "PodGroup", definition: "io.k8s.api.scheduling.v1alpha2.PodGroup"},
		}},
	}

	manifestPath := filepath.Join("..", "..", "kubernetesapi", "builtin-versions.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var manifest versionManifest
	require.NoError(t, json.Unmarshal(manifestBytes, &manifest))
	require.NoError(t, validateManifest(&manifest))

	expectedVersions := make([]string, 0, len(expected))
	for _, minor := range expected {
		expectedVersions = append(expectedVersions, minor.kubernetesVersion)
	}
	require.GreaterOrEqual(t, len(manifest.Versions), len(expectedVersions))
	require.Equal(t, expectedVersions, manifest.Versions[:len(expectedVersions)])

	bundlePath := filepath.Join("..", "..", "kubernetesapi", "data",
		"kubernetes-openapi-union.bundle-v1.json.gz")
	bundleBytes, err := readInput(bundlePath)
	require.NoError(t, err)
	var bundle builtinopenapi.Bundle
	require.NoError(t, json.Unmarshal(bundleBytes, &bundle))
	require.NoError(t, bundle.Validate())
	bundleVersions := make([]string, len(bundle.Sources))
	for i, source := range bundle.Sources {
		bundleVersions[len(bundle.Sources)-1-i] = source.KubernetesVersion
	}
	require.Equal(t, manifest.Versions, bundleVersions)
	bundleResources := make(map[string]builtinopenapi.Resource, len(bundle.Resources))
	for _, resource := range bundle.Resources {
		bundleResources[resourceKey(resource.APIVersion, resource.Kind)] = resource
	}

	for _, minor := range expected {
		t.Run(minor.kubernetesVersion, func(t *testing.T) {
			require.NotEmpty(t, minor.apis, "each covered minor must have an API provenance check")
			for _, api := range minor.apis {
				t.Run(api.apiVersion+"/"+api.kind, func(t *testing.T) {
					key := resourceKey(api.apiVersion, api.kind)
					resource, found := bundleResources[key]
					require.True(t, found, "API is missing from the compiled union")
					require.Equal(t, api.definition, resource.Definition)
					require.Contains(t, bundle.Definitions, api.definition)
				})
			}
		})
	}
}

func TestOpenAPICorrectionExceptionIsIncluded(t *testing.T) {
	bundleBytes, err := readInput(filepath.Join("..", "..", "kubernetesapi", "data",
		"kubernetes-openapi-union.bundle-v1.json.gz"))
	require.NoError(t, err)
	var bundle builtinopenapi.Bundle
	require.NoError(t, json.Unmarshal(bundleBytes, &bundle))

	status := bundle.Definitions["io.k8s.api.resource.v1alpha1.ResourceClaimStatus"]
	reservedFor := status.Properties["reservedFor"]
	require.Equal(t, "map", reservedFor.Extensions["x-kubernetes-list-type"])
	require.Equal(t, []interface{}{"uid"}, reservedFor.Extensions["x-kubernetes-list-map-keys"])
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

func TestCompileSourcesUsesNewestSchemaAndNormalizesGVKUnion(t *testing.T) {
	newest := sourceDocument{
		kubernetesVersion: "v1.36.0",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.36.0"},
  "paths": {},
  "definitions": {
    "shared": {"type": "object", "properties": {"newest": {"type": "string"}}},
    "resource": {
      "type": "object",
      "description": "newest schema",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1", "kind": "Current"}],
      "properties": {
        "newest": {"type": "string"},
        "spec": {"$ref": "#/definitions/shared"}
      }
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
    "resource": {
      "type": "object",
      "description": "oldest schema",
      "x-kubernetes-group-version-kind": [{"group": "example.io", "version": "v1beta1", "kind": "Legacy"}],
      "properties": {
        "oldest": {"type": "integer"},
        "spec": {"$ref": "#/definitions/shared"}
      }
    }
  }
}`),
	}

	bundle, err := compileSources([]sourceDocument{newest, oldest}, builtinopenapi.Coverage{
		Floor: "v1.21.2", Ceiling: "v1.36.0",
	})
	require.NoError(t, err)
	require.Equal(t, []builtinopenapi.Resource{
		{APIVersion: "example.io/v1", Kind: "Current", Definition: "resource"},
		{APIVersion: "example.io/v1beta1", Kind: "Legacy", Definition: "resource"},
	}, bundle.Resources)

	shared := bundle.Definitions["shared"]
	require.Contains(t, shared.Properties, "newest")
	require.NotContains(t, shared.Properties, "oldest", "definitions must not be merged field-by-field")

	resource := bundle.Definitions["resource"]
	require.Equal(t, "newest schema", resource.Description)
	require.Contains(t, resource.Properties, "newest")
	require.NotContains(t, resource.Properties, "oldest", "schema fields must come from the newest source")
	require.Equal(t, []interface{}{
		map[string]interface{}{"group": "example.io", "version": "v1", "kind": "Current"},
		map[string]interface{}{"group": "example.io", "version": "v1beta1", "kind": "Legacy"},
	}, resource.Extensions[gvkExtension], "GVKs must be normalized from the union resource inventory")
}

func TestNormalizeDefinitionGVKsUsesResourceInventory(t *testing.T) {
	resourceSchema := spec.Schema{}
	resourceSchema.Extensions = spec.Extensions{
		gvkExtension: []interface{}{
			map[string]interface{}{"group": "stale.example", "version": "v1", "kind": "Stale"},
		},
		"x-preserved": "value",
	}
	unusedSchema := spec.Schema{}
	unusedSchema.Extensions = spec.Extensions{
		gvkExtension: []interface{}{
			map[string]interface{}{"group": "stale.example", "version": "v1", "kind": "Unused"},
		},
	}
	definitions := spec.Definitions{
		"resource": resourceSchema,
		"unused":   unusedSchema,
	}
	resources := []builtinopenapi.Resource{
		{APIVersion: "example.io/v1", Kind: "PathOnly"},
		{APIVersion: "v1", Kind: "Pod", Definition: "resource"},
	}

	require.NoError(t, normalizeDefinitionGVKs(definitions, resources))
	resourceSchema = definitions["resource"]
	require.Equal(t, "value", resourceSchema.Extensions["x-preserved"])
	require.Equal(t, []interface{}{
		map[string]interface{}{"group": "", "version": "v1", "kind": "Pod"},
	}, resourceSchema.Extensions[gvkExtension])
	require.NotContains(t, definitions["unused"].Extensions, gvkExtension)

	err := normalizeDefinitionGVKs(definitions, []builtinopenapi.Resource{{
		APIVersion: "apps/v1", Kind: "Deployment", Definition: "missing",
	}})
	require.ErrorContains(t, err, `references missing definition "missing"`)
}

func TestCompileSourcesValidatesEachSourceReferences(t *testing.T) {
	broken := sourceDocument{
		kubernetesVersion: "v1.36.0",
		input: []byte(`{
  "swagger": "2.0",
  "info": {"title": "test", "version": "v1.36.0"},
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
		Floor: "v1.21.2", Ceiling: "v1.36.0",
	})
	require.ErrorContains(t, err, "validate Kubernetes v1.36.0 definitions")
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
	newest := source("v1.36.0", "/apis/example.io/v1/namespaces/{namespace}/widgets")
	oldest := source("v1.21.2", "/apis/example.io/v1/widgets")

	_, err := compileSources([]sourceDocument{newest, oldest}, builtinopenapi.Coverage{
		Floor: "v1.21.2", Ceiling: "v1.36.0",
	})
	require.ErrorContains(t, err, `has scopes "Namespaced" and "Cluster"`)
}

func TestParseGVKRejectsNonStringGroup(t *testing.T) {
	_, _, err := parseGVK(map[string]interface{}{
		"group":   1,
		"version": "v1",
		"kind":    "Pod",
	})
	require.ErrorContains(t, err, "group must be a string")
}

func TestValidateManifest(t *testing.T) {
	exception := func(reason, upstreamPullRequest string) patchVersionException {
		return patchVersionException{
			Reason:              reason,
			UpstreamPullRequest: upstreamPullRequest,
		}
	}
	tests := []struct {
		name       string
		versions   []string
		exceptions map[string]patchVersionException
		wantErr    string
	}{
		{name: "valid", versions: []string{"v1.35.0", "v1.36.0"}},
		{name: "valid OpenAPI correction", versions: []string{"v1.35.0", "v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionOpenAPICorrection, "https://github.com/kubernetes/kubernetes/pull/123"),
		}},
		{name: "valid legacy baseline", versions: []string{"v1.35.2", "v1.36.0"}, exceptions: map[string]patchVersionException{
			"v1.35.2": exception(patchVersionExceptionLegacyBaseline, ""),
		}},
		{name: "out of order", versions: []string{"v1.36.0", "v1.35.0"}, wantErr: "oldest-first"},
		{name: "duplicate", versions: []string{"v1.36.0", "v1.36.0"}, wantErr: "oldest-first"},
		{name: "duplicate minor", versions: []string{"v1.36.0", "v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionOpenAPICorrection, "https://github.com/kubernetes/kubernetes/pull/123"),
		}, wantErr: "exactly one snapshot"},
		{name: "minor gap", versions: []string{"v1.34.0", "v1.36.0"}, wantErr: "exactly one snapshot"},
		{name: "non-zero patch without exception", versions: []string{"v1.36.1"}, wantErr: "requires patchVersionException"},
		{name: "exception on zero patch", versions: []string{"v1.36.0"}, exceptions: map[string]patchVersionException{
			"v1.36.0": exception(patchVersionExceptionOpenAPICorrection, "https://github.com/kubernetes/kubernetes/pull/123"),
		}, wantErr: "only valid for a non-zero patch"},
		{name: "legacy baseline is not oldest", versions: []string{"v1.35.0", "v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionLegacyBaseline, ""),
		}, wantErr: "only valid for the oldest source"},
		{name: "legacy baseline has pull request", versions: []string{"v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionLegacyBaseline, "https://github.com/kubernetes/kubernetes/pull/123"),
		}, wantErr: "must not specify upstreamPullRequest"},
		{name: "OpenAPI correction missing pull request", versions: []string{"v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionOpenAPICorrection, ""),
		}, wantErr: "requires upstreamPullRequest"},
		{name: "OpenAPI correction invalid pull request", versions: []string{"v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionOpenAPICorrection, "https://github.com/kubernetes/kubernetes/issues/123"),
		}, wantErr: "requires upstreamPullRequest"},
		{name: "OpenAPI correction pull request without URL", versions: []string{"v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception(patchVersionExceptionOpenAPICorrection, "123"),
		}, wantErr: "requires upstreamPullRequest"},
		{name: "unknown exception reason", versions: []string{"v1.36.1"}, exceptions: map[string]patchVersionException{
			"v1.36.1": exception("latest-patch", ""),
		}, wantErr: "unsupported patchVersionException reason"},
		{name: "exception for unselected version", versions: []string{"v1.36.0"}, exceptions: map[string]patchVersionException{
			"v1.35.1": exception(patchVersionExceptionOpenAPICorrection, "https://github.com/kubernetes/kubernetes/pull/123"),
		}, wantErr: "is not selected"},
		{name: "invalid version", versions: []string{"1.36"}, wantErr: "invalid Kubernetes version"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manifest := versionManifest{Versions: tc.versions, PatchVersionExceptions: tc.exceptions}
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

func TestReadInputWithLimit(t *testing.T) {
	const limit = int64(4)
	testDir := t.TempDir()

	write := func(t *testing.T, name string, contents []byte) string {
		t.Helper()
		path := filepath.Join(testDir, name)
		require.NoError(t, os.WriteFile(path, contents, 0o600))
		return path
	}
	gzipContents := func(t *testing.T, contents []byte) []byte {
		t.Helper()
		var buffer bytes.Buffer
		writer := gzip.NewWriter(&buffer)
		_, err := writer.Write(contents)
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		return buffer.Bytes()
	}

	t.Run("raw at limit", func(t *testing.T) {
		path := write(t, "raw", []byte("data"))
		got, err := readInputWithLimit(path, limit)
		require.NoError(t, err)
		require.Equal(t, []byte("data"), got)
	})

	t.Run("raw exceeds limit", func(t *testing.T) {
		path := write(t, "raw-large", []byte("large"))
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "input")
		require.ErrorContains(t, err, "exceeds 4 bytes")
	})

	t.Run("gzip at limit", func(t *testing.T) {
		path := write(t, "gzip", gzipContents(t, []byte("data")))
		got, err := readInputWithLimit(path, limit)
		require.NoError(t, err)
		require.Equal(t, []byte("data"), got)
	})

	t.Run("gzip exceeds limit", func(t *testing.T) {
		path := write(t, "gzip-large", gzipContents(t, []byte("large")))
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "decompressed input")
		require.ErrorContains(t, err, "exceeds 4 bytes")
	})

	t.Run("invalid gzip header", func(t *testing.T) {
		path := write(t, "gzip-invalid-header", []byte{0x1f, 0x8b})
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "open gzip input")
	})

	t.Run("invalid gzip body", func(t *testing.T) {
		contents := gzipContents(t, []byte("data"))
		contents[len(contents)-1]++
		path := write(t, "gzip-invalid-body", contents)
		_, err := readInputWithLimit(path, limit)
		require.ErrorContains(t, err, "read decompressed input")
	})
}
