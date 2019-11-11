// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kustfile

import (
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	testutils_test "sigs.k8s.io/kustomize/kustomize/v3/internal/commands/testutils"
)

func TestFieldOrder(t *testing.T) {
	expected := []string{
		"APIVersion",
		"Kind",
		"Resources",
		"Bases",
		"NamePrefix",
		"NameSuffix",
		"Namespace",
		"Crds",
		"CommonLabels",
		"CommonAnnotations",
		"PatchesStrategicMerge",
		"PatchesJson6902",
		"Patches",
		"ConfigMapGenerator",
		"SecretGenerator",
		"GeneratorOptions",
		"Vars",
		"Images",
		"Replicas",
		"Configurations",
		"Generators",
		"Transformers",
		"Inventory",
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
	kustomization.FixKustomizationPostUnmarshalling()
	if !reflect.DeepEqual(kustomization, content) {
		t.Fatal("Read kustomization is different from written kustomization")
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
		fSys.WriteFile(n, []byte(kcontent))
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
  immediateSubstitution: false
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
	if !reflect.DeepEqual(kustomizationContentWithComments, bytes) {
		t.Fatal("written kustomization with comments is not the same as original one")
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
  immediateSubstitution: false
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

	if string(expected) != string(bytes) {
		t.Fatalf(
			"expected =\n%s\n\nactual =\n%s\n",
			string(expected), string(bytes))
	}
}

func TestFixPatchesField(t *testing.T) {
	kustomizationContentWithComments := []byte(`
patches:
- patch1.yaml
- patch2.yaml
`)

	expected := []byte(`
patchesStrategicMerge:
- patch1.yaml
- patch2.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
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

	if string(expected) != string(bytes) {
		t.Fatalf(
			"expected =\n%s\n\nactual =\n%s\n",
			string(expected), string(bytes))
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

	if string(expected) != string(bytes) {
		t.Fatalf(
			"expected =\n%s\n\nactual =\n%s\n",
			string(expected), string(bytes))
	}
}
