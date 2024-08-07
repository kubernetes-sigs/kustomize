// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kustfile

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v5/commands/internal/testutils"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func TestFieldOrder(t *testing.T) {
	expected := []string{
		"APIVersion",
		"Kind",
		"MetaData",
		"SortOptions",
		"Resources",
		"Bases",
		"NamePrefix",
		"NameSuffix",
		"Namespace",
		"Crds",
		"CommonLabels",
		"Labels",
		"CommonAnnotations",
		"PatchesStrategicMerge",
		"PatchesJson6902",
		"Patches",
		"ConfigMapGenerator",
		"SecretGenerator",
		"HelmCharts",
		"HelmChartInflationGenerator",
		"HelmGlobals",
		"GeneratorOptions",
		"Vars",
		"Images",
		"Replacements",
		"Replicas",
		"Configurations",
		"Generators",
		"Transformers",
		"Validators",
		"Components",
		"OpenAPI",
		"BuildMetadata",
	}
	actual := determineFieldOrder()
	if len(expected) != len(actual) {
		t.Fatalf("Incorrect field count.")
	}
	for i, n := range expected {
		if n != actual[i] {
			t.Fatalf("Bad field order.")
		}
	}
}

func TestWriteAndRead(t *testing.T) {
	kustomization := &types.Kustomization{
		NamePrefix: "prefix",
	}

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if err := mf.Write(kustomization); err != nil {
		t.Fatalf("Couldn't write kustomization file: %v\n", err)
	}

	content, err := mf.Read()
	if err != nil {
		t.Fatalf("Couldn't read kustomization file: %v\n", err)
	}
	kustomization.FixKustomization()
	if !reflect.DeepEqual(kustomization, content) {
		t.Fatal("Read kustomization is different from written kustomization")
	}
}

func TestReadAndWrite(t *testing.T) {
	kWrite := []byte(completeKustfileInOrder)

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kWrite)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	kustomization, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if err := mf.Write(kustomization); err != nil {
		t.Fatalf("Couldn't write kustomization file: %v\n", err)
	}

	kRead, err := testutils_test.ReadTestKustomization(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if !reflect.DeepEqual(kWrite, kRead) {
		t.Fatal("Written kustomization is different from read kustomization")
	}
}

func TestReadAndWriteDummy(t *testing.T) {
	kWrite := &types.Kustomization{
		TypeMeta: types.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
		},
		MetaData: &types.ObjectMeta{
			Name:        "name",
			Namespace:   "namespace",
			Labels:      map[string]string{"label": "label"},
			Annotations: map[string]string{"annotation": "annotation"},
		},
		OpenAPI:      map[string]string{"path": "schema.json"},
		NamePrefix:   "prefix",
		NameSuffix:   "suffix",
		Namespace:    "namespace",
		CommonLabels: map[string]string{"commonLabel": "commonLabel"},
		Labels: []types.Label{{
			Pairs:            map[string]string{"label": "label"},
			IncludeSelectors: true,
			IncludeTemplates: true,
			FieldSpecs: []types.FieldSpec{{
				Path: "metadata.labels.label",
			}},
		}},
		CommonAnnotations: map[string]string{"commonAnnotation": "commonAnnotation"},
		Patches: []types.Patch{{
			Path: "path",
		}},
		Images: []types.Image{{
			Name:      "name",
			NewName:   "newName",
			TagSuffix: "tagSuffix",
			NewTag:    "newTag",
			Digest:    "digest",
		}},
		Replacements: []types.ReplacementField{{
			Path: "path",
		}},
		Replicas: []types.Replica{{
			Name:  "name",
			Count: 1,
		}},
		SortOptions: &types.SortOptions{
			Order: types.LegacySortOrder,
			LegacySortOptions: &types.LegacySortOptions{
				OrderFirst: []string{"orderFirst"},
				OrderLast:  []string{"orderLast"},
			},
		},
		Resources:  []string{"resource"},
		Components: []string{"component"},
		Crds:       []string{"crd"},
		ConfigMapGenerator: []types.ConfigMapArgs{{
			GeneratorArgs: types.GeneratorArgs{
				Namespace: "namespace",
				Name:      "name",
			},
		}},
		SecretGenerator: []types.SecretArgs{{
			GeneratorArgs: types.GeneratorArgs{
				Namespace: "namespace",
				Name:      "name",
			},
		}},
		HelmGlobals: &types.HelmGlobals{
			ChartHome:  "chartHome",
			ConfigHome: "configHome",
		},
		HelmCharts: []types.HelmChart{{
			Name: "name",
		}},
		GeneratorOptions: &types.GeneratorOptions{
			Labels: map[string]string{"label": "label"},
		},
		Configurations: []string{"configuration"},
		Generators:     []string{"generator"},
		Transformers:   []string{"transformer"},
		Validators:     []string{"validator"},
		BuildMetadata:  []string{"buildMetadata"},
	}

	// this check is for forward compatibility: if this fails, add a dummy value to the Kustomization above
	assertAllNonZeroExcept(t, kWrite, []string{"PatchesStrategicMerge", "PatchesJson6902", "ImageTags", "Vars", "Bases", "HelmChartInflationGenerator"})

	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomization(fSys)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if err := mf.Write(kWrite); err != nil {
		t.Fatalf("Couldn't write kustomization file: %v\n", err)
	}

	kRead, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if !reflect.DeepEqual(kWrite, kRead) {
		t.Fatal("Written kustomization is different from read kustomization.")
	}
}

func assertAllNonZeroExcept(t *testing.T, val *types.Kustomization, except []string) {
	t.Helper()
	fFor := reflect.ValueOf(val).Elem()
	n := fFor.NumField()
	for i := 0; i < n; i++ {
		key := fFor.Type().Field(i).Name
		val := fFor.Field(i)
		if val.IsZero() && !slices.Contains(except, key) {
			t.Helper()
			t.Fatalf("Key %s should not be empty", key)
		}
	}
}

func TestGetPath(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	testutils_test.WriteTestKustomization(fSys)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if mf.GetPath() != "kustomization.yaml" {
		t.Fatalf("Path expected: kustomization.yaml. Actual path: %v", mf.GetPath())
	}
}

func TestNewNotExist(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	_, err := NewKustomizationFile(fSys)
	if err == nil {
		t.Fatalf("expect an error")
	}
	contained := "Missing kustomization file"
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
	_, err = NewKustomizationFile(fSys)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
}

func TestAllKustomizationFileNames(t *testing.T) {
	kcontent := `
configMapGenerator:
- literals:
  - foo=bar
  - baz=qux
  name: my-configmap
`
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		fSys := filesys.MakeFsInMemory()
		err := fSys.WriteFile(n, []byte(kcontent))
		require.NoError(t, err)
		k, err := NewKustomizationFile(fSys)
		if err != nil {
			t.Fatalf("Unexpected Error: %v", err)
		}
		if k.path != n {
			t.Fatalf("Load incorrect file path %s", k.path)
		}
	}
}

func TestPreserveComments(t *testing.T) {
	kustomizationContentWithComments := []byte(
		`# shem qing some comments
# This is some comment we should preserve
# don't delete it
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../namespaces
- pod.yaml
- service.yaml
# something you may want to keep
vars:
- fieldref:
    fieldPath: metadata.name
  name: MY_SERVICE_NAME
  objref:
    apiVersion: v1
    kind: Service
    name: my-service
# some descriptions for the patches
patchesStrategicMerge:
- service.yaml
- pod.yaml
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationContentWithComments)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	kustomization, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if err = mf.Write(kustomization); err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	bytes, _ := fSys.ReadFile(mf.path)

	if diff := cmp.Diff(kustomizationContentWithComments, bytes); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}

func TestPreserveCommentsWithAdjust(t *testing.T) {
	kustomizationContentWithComments := []byte(`



# Some comments
# This is some comment we should preserve
# don't delete it
RESOURCES:
- ../namespaces
- pod.yaml
  # See which field this comment goes into
- service.yaml

apiVersion: kustomize.config.k8s.io/v1beta1
kind: kustomization

# something you may want to keep
vars:
- fieldref:
    fieldPath: metadata.name
  name: MY_SERVICE_NAME
  objref:
    apiVersion: v1
    kind: Service
    name: my-service

# some descriptions for the patches

patchesStrategicMerge:
- service.yaml
- pod.yaml
# generator options
generatorOptions:
  disableNameSuffixHash: true
`)

	expected := []byte(`



# Some comments
# This is some comment we should preserve
# don't delete it
  # See which field this comment goes into
resources:
- ../namespaces
- pod.yaml
- service.yaml

apiVersion: kustomize.config.k8s.io/v1beta1
kind: kustomization

# something you may want to keep
vars:
- fieldref:
    fieldPath: metadata.name
  name: MY_SERVICE_NAME
  objref:
    apiVersion: v1
    kind: Service
    name: my-service

# some descriptions for the patches

patchesStrategicMerge:
- service.yaml
- pod.yaml
# generator options
generatorOptions:
  disableNameSuffixHash: true
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(
		fSys, kustomizationContentWithComments)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	kustomization, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if err = mf.Write(kustomization); err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	bytes, _ := fSys.ReadFile(mf.path)

	if diff := cmp.Diff(expected, bytes); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}

func TestFixPatchesFieldForExtendedPatch(t *testing.T) {
	kustomizationContentWithComments := []byte(`
patches:
- path: patch1.yaml
  target:
    kind: Deployment
- path: patch2.yaml
  target:
    kind: Service
`)

	expected := []byte(`
patches:
- path: patch1.yaml
  target:
    kind: Deployment
- path: patch2.yaml
  target:
    kind: Service
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kustomizationContentWithComments)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	kustomization, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if err = mf.Write(kustomization); err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	bytes, _ := fSys.ReadFile(mf.path)

	if diff := cmp.Diff(expected, bytes); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}

func TestCommentsWithDocumentSeperatorAtBeginning(t *testing.T) {
	kustomizationContentWithComments := []byte(`


# Some comments
# This is some comment we should preserve
# don't delete it
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: mynamespace
`)

	expected := []byte(`


# Some comments
# This is some comment we should preserve
# don't delete it
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: mynamespace
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(
		fSys, kustomizationContentWithComments)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	kustomization, err := mf.Read()
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if err = mf.Write(kustomization); err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	bytes, _ := fSys.ReadFile(mf.path)

	if diff := cmp.Diff(expected, bytes); diff != "" {
		t.Errorf("Mismatch (-expected, +actual):\n%s", diff)
	}
}

func TestUnknownFieldInKustomization(t *testing.T) {
	kContent := []byte(`
foo:
  bar
`)
	fSys := filesys.MakeFsInMemory()
	testutils_test.WriteTestKustomizationWith(fSys, kContent)
	mf, err := NewKustomizationFile(fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	_, err = mf.Read()
	if err == nil || err.Error() != "invalid Kustomization: json: unknown field \"foo\"" {
		t.Fatalf("Expect an unknown field error but got: %v", err)
	}
}

const completeKustfileInOrder = `
kind: Kustomization
apiVersion: kustomize.config.k8s.io/v1beta1
metadata:
  annotations:
    annotation: annotation
  labels:
    label: label
  name: name
  namespace: namespace
openapi:
  path: schema.json
namePrefix: prefix
nameSuffix: suffix
namespace: namespace
commonLabels:
  commonLabel: commonLabel
labels:
- fields:
  - path: metadata.labels.label
  includeSelectors: true
  includeTemplates: true
  pairs:
    label: label
commonAnnotations:
  commonAnnotation: commonAnnotation
patches:
- path: path
images:
- digest: digest
  name: name
  newName: newName
  newTag: newTag
  tagSuffix: tagSuffix
replacements:
- path: path
replicas:
- count: 1
  name: name
sortOptions:
  legacySortOptions:
    orderFirst:
    - orderFirst
    orderLast:
    - orderLast
  order: legacy
resources:
- resource
components:
- component
crds:
- crd
configMapGenerator:
- name: name
  namespace: namespace
secretGenerator:
- name: name
  namespace: namespace
helmGlobals:
  chartHome: chartHome
  configHome: configHome
helmCharts:
- name: name
- name: chartName
generatorOptions:
  labels:
    label: label
configurations:
- configuration
generators:
- generator
transformers:
- transformer
buildMetadata:
- buildMetadata
`
