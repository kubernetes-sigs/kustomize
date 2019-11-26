// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kustomize/v3/internal/commands/kustfile"
)

var factory = kunstruct.NewKunstructuredFactoryImpl()

func readKustomizationFS(t *testing.T, fSys filesys.FileSystem) *types.Kustomization {
	kf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		t.Errorf("unexpected new error %v", err)
	}
	m, err := kf.Read()
	if err != nil {
		t.Errorf("unexpected read error %v", err)
	}
	return m
}
func TestCreateNoArgs(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	cmd := NewCmdCreate(fSys, factory)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	readKustomizationFS(t, fSys)
}

func TestCreateWithResources(t *testing.T) {
	fSys := filesys.MakeEmptyDirInMemory()
	fSys.WriteFile("foo.yaml", []byte(""))
	fSys.WriteFile("bar.yaml", []byte(""))
	opts := createFlags{resources: "foo.yaml,bar.yaml"}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	expected := []string{"foo.yaml", "bar.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
	}
}

func TestCreateWithNamespace(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	want := "foo"
	opts := createFlags{namespace: want}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	got := m.Namespace
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestCreateWithLabels(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	opts := createFlags{labels: "foo:bar"}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	expected := map[string]string{"foo": "bar"}
	if !reflect.DeepEqual(m.CommonLabels, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.CommonLabels)
	}
}

func TestCreateWithAnnotations(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	opts := createFlags{annotations: "foo:bar"}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	expected := map[string]string{"foo": "bar"}
	if !reflect.DeepEqual(m.CommonAnnotations, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.CommonAnnotations)
	}
}

func TestCreateWithNamePrefix(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	want := "foo-"
	opts := createFlags{prefix: want}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	got := m.NamePrefix
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestCreateWithNameSuffix(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	opts := createFlags{suffix: "-foo"}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	if m.NameSuffix != "-foo" {
		t.Errorf("want: -foo, got: %s", m.NameSuffix)
	}
}

func writeDetectContent(fSys filesys.FileSystem) {
	fSys.WriteFile("/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test`))
	fSys.WriteFile("/README.md", []byte(`
# Not a k8s resource
This file is not a valid kubernetes object.`))
	fSys.WriteFile("/non-k8s.yaml", []byte(`
# Not a k8s resource
other: yaml
foo:
- bar
- baz`))
	fSys.Mkdir("/sub")
	fSys.WriteFile("/sub/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test2`))
	fSys.WriteFile("/sub/README.md", []byte(`
# Not a k8s resource
This file in a subdirectory is not a valid kubernetes object.`))
	fSys.WriteFile("/sub/non-k8s.yaml", []byte(`
# Not a k8s resource
other: yaml
foo:
- bar
- baz`))
	fSys.Mkdir("/overlay")
	fSys.WriteFile("/overlay/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test3`))
	fSys.WriteFile("/overlay/kustomization.yaml", []byte(`
resources:
- test.yaml`))
}

func TestCreateWithDetect(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	writeDetectContent(fSys)
	opts := createFlags{path: "/", detectResources: true}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	expected := []string{"/test.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
	}
}

func TestCreateWithDetectRecursive(t *testing.T) {
	fSys := filesys.MakeFsInMemory()
	writeDetectContent(fSys)
	opts := createFlags{path: "/", detectResources: true, detectRecursive: true}
	err := runCreate(opts, fSys, factory)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fSys)
	expected := []string{"/overlay", "/sub/test.yaml", "/test.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
	}
}
