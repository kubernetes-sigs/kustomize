/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kustfile

import (
	"path"
	"reflect"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/types"
)

const (
	curDir = "./"
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
		"ConfigMapGenerator",
		"SecretGenerator",
		"GeneratorOptions",
		"Vars",
		"ImageTags",
		"Configurations",
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

func TestReadExisting(t *testing.T) {
	readFromExistingFile := func(dir string, filename string) {
		kustomizationContent := []byte(`
apiVersion: v1beta1
kind: Kustomization
namespace: my-recognizable-namespace
`)

		fSys := fs.MakeFakeFS()

		err := fSys.MkdirAll(dir)
		if err != nil {
			t.Fatalf("Test bench failure: Couldn't MkdirAll(): %v\n", err)
		}

		targetPath := path.Join(dir, filename)
		err = fSys.WriteFile(targetPath, kustomizationContent)
		if err != nil {
			t.Fatalf("Test bench failure: Couldn't WriteFile() at %v: %v\n", targetPath, err)
		}

		mf, err := NewKustomizationFile(dir, fSys)
		if err != nil {
			t.Fatalf("Couldn't NewKustomizationFile() at %v: %v\n", targetPath, err)
		}

		content, err := mf.Read()
		if err != nil {
			t.Fatalf("Couldn't Read() at %v: %v\n", targetPath, err)
		}

		if content.Namespace != "my-recognizable-namespace" {
			t.Fatalf(
				"Read kustomization file at %v but it didn't contain the correct content\n",
				targetPath,
			)
		}
	}

	// Test preferred name
	readFromExistingFile("./", "kustomization.yaml")
	readFromExistingFile("../", "kustomization.yaml")
	readFromExistingFile("/", "kustomization.yaml")
	readFromExistingFile("subdir", "kustomization.yaml")
	readFromExistingFile("subdir1/subdir2", "kustomization.yaml")

	// Test fallback name
	readFromExistingFile("./", "kustomization.yml")
	readFromExistingFile("../", "kustomization.yml")
	readFromExistingFile("/", "kustomization.yml")
	readFromExistingFile("subdir", "kustomization.yml")
	readFromExistingFile("subdir1/subdir2", "kustomization.yml")
}

func TestReadReflectsWrite(t *testing.T) {
	readReflectsWrite := func(dir string, filename string) {
		kustomizationContent := []byte(`
apiVersion: v1beta1
kind: Kustomization
namespace: my-recognizable-namespace-1
`)

		newKustomization := &types.Kustomization{
			Namespace: "my-recognizable-namespace-2",
		}

		fSys := fs.MakeFakeFS()

		err := fSys.MkdirAll(dir)
		if err != nil {
			t.Fatalf("Test bench failure: Couldn't MkdirAll(): %v\n", err)
		}

		targetPath := path.Join(dir, filename)
		err = fSys.WriteFile(targetPath, kustomizationContent)
		if err != nil {
			t.Fatalf("Test bench failure: Couldn't WriteFile() at %v: %v\n", targetPath, err)
		}

		mf, err := NewKustomizationFile(dir, fSys)
		if err != nil {
			t.Fatalf("Couldn't NewKustomizationFile() at %v: %v\n", targetPath, err)
		}

		err = mf.Write(newKustomization)
		if err != nil {
			t.Fatalf("Couldn't Write() at %v: %v\n", targetPath, err)
		}

		content, err := mf.Read()
		if err != nil {
			t.Fatalf("Couldn't Read() at %v: %v\n", targetPath, err)
		}

		if content.Namespace != "my-recognizable-namespace-2" {
			t.Fatalf(
				"Read kustomization file at %v but it didn't contain the correct content\n",
				targetPath,
			)
		}
	}

	// Test preferred name
	readReflectsWrite("./", "kustomization.yaml")
	readReflectsWrite("../", "kustomization.yaml")
	readReflectsWrite("/", "kustomization.yaml")
	readReflectsWrite("subdir", "kustomization.yaml")
	readReflectsWrite("subdir1/subdir2", "kustomization.yaml")

	// Test fallback name
	readReflectsWrite("./", "kustomization.yml")
	readReflectsWrite("../", "kustomization.yml")
	readReflectsWrite("/", "kustomization.yml")
	readReflectsWrite("subdir", "kustomization.yml")
	readReflectsWrite("subdir1/subdir2", "kustomization.yml")
}

// Deprecated fields should not survive being read.
func TestDeprecationOfPatches(t *testing.T) {
	hasDeprecatedFields := []byte(`
namePrefix: acme
nameSuffix: emca
patches:
- alice
patchesStrategicMerge:
- bob
`)
	fSys := fs.MakeFakeFS()
	fSys.WriteTestKustomizationWith(hasDeprecatedFields)
	mf, err := NewKustomizationFile(curDir, fSys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	k, err := mf.Read()
	if err != nil {
		t.Fatalf("Couldn't read kustomization file: %v\n", err)
	}
	if k.NamePrefix != "acme" {
		t.Fatalf("Unexpected name prefix")
	}
	if k.NameSuffix != "emca" {
		t.Fatalf("Unexpected name suffix")
	}
	if len(k.Patches) > 0 {
		t.Fatalf("Expected nothing in Patches.")
	}
	if len(k.PatchesStrategicMerge) != 2 {
		t.Fatalf(
			"Expected len(k.PatchesStrategicMerge) == 2, got %d",
			len(k.PatchesStrategicMerge))
	}
	m := make(map[string]bool)
	for _, v := range k.PatchesStrategicMerge {
		m[string(v)] = true
	}
	if _, f := m["alice"]; !f {
		t.Fatalf("Expected alice in PatchesStrategicMerge")
	}
	if _, f := m["bob"]; !f {
		t.Fatalf("Expected bob in PatchesStrategicMerge")
	}
}

func TestNewNotExist(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	_, err := NewKustomizationFile("./", fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	contained := "Missing kustomization file"
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
	_, err = NewKustomizationFile("./", fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
}

func TestSecondarySuffix(t *testing.T) {
	kcontent := `
configMapGenerator:
- literals:
  - foo=bar
  - baz=qux
  name: my-configmap
`
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile(constants.SecondaryKustomizationFileName, []byte(kcontent))
	k, err := NewKustomizationFile("./", fakeFS)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}
	if k.path != constants.SecondaryKustomizationFileName {
		t.Fatalf("Load incorrect file path %s", k.path)
	}
}

func TestPreserveComments(t *testing.T) {
	kustomizationContentWithComments := []byte(
		`# shem qing some comments
# This is some comment we should preserve
# don't delete it
apiVersion: v1beta1
kind: Kustomization
resources:
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
bases:
- ../namespaces
# some descriptions for the patches
patchesStrategicMerge:
- service.yaml
- pod.yaml
`)
	fSys := fs.MakeFakeFS()
	fSys.WriteTestKustomizationWith(kustomizationContentWithComments)
	mf, err := NewKustomizationFile("./", fSys)
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

    

# shem qing some comments
# This is some comment we should preserve
# don't delete it
resources:
- pod.yaml
  # See which field this comment goes into
- service.yaml

APIVersion: v1beta1
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

BASES:
- ../namespaces

# some descriptions for the patches

patchesStrategicMerge:
- service.yaml
- pod.yaml
# generator options
generatorOptions:
  disableNameSuffixHash: true
`)

	expected := []byte(`

    

# shem qing some comments
# This is some comment we should preserve
# don't delete it
  # See which field this comment goes into
resources:
- pod.yaml
- service.yaml

apiVersion: v1beta1
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

bases:
- ../namespaces

# some descriptions for the patches

patchesStrategicMerge:
- service.yaml
- pod.yaml
# generator options
generatorOptions:
  disableNameSuffixHash: true
`)
	fSys := fs.MakeFakeFS()
	fSys.WriteTestKustomizationWith(kustomizationContentWithComments)
	mf, err := NewKustomizationFile("./", fSys)
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

	if !reflect.DeepEqual(expected, bytes) {
		t.Fatal("written kustomization with comments is not the same as original one\n", string(bytes))
	}
}
