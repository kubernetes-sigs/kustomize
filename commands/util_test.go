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

package commands

import (
	"reflect"
	"strings"
	"testing"

	manifest "k8s.io/kubectl/pkg/apis/manifest/v1alpha1"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

func TestWriteAndRead(t *testing.T) {
	manifest := &manifest.Manifest{
		NamePrefix: "prefix",
	}

	fsys := fs.MakeFakeFS()
	fsys.Create("kustomize.yaml")
	mf, err := newManifestFile("kustomize.yaml", fsys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if err := mf.write(manifest); err != nil {
		t.Fatalf("Couldn't write manifest file: %v\n", err)
	}

	readManifest, err := mf.read()
	if err != nil {
		t.Fatalf("Couldn't read manifest file: %v\n", err)
	}
	if !reflect.DeepEqual(manifest, readManifest) {
		t.Fatal("Read manifest is different from written manifest")
	}
}

func TestEmptyFile(t *testing.T) {
	fsys := fs.MakeFakeFS()
	_, err := newManifestFile("", fsys)
	if err == nil {
		t.Fatalf("Creat manifestFile from empty filename should fail")
	}
}

func TestNewNotExist(t *testing.T) {
	badSuffix := "foo.bar"
	fakeFS := fs.MakeFakeFS()
	fakeFS.Mkdir(".", 0644)
	fakeFS.Create(badSuffix)
	_, err := newManifestFile("kustomize.yaml", fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), "Run `kustomize init` first") {
		t.Fatalf("expect an error contains %q, but got %v", "does not exist", err)
	}
	_, err = newManifestFile("kustomize.yaml", fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), "Run `kustomize init` first") {
		t.Fatalf("expect an error contains %q, but got %v", "does not exist", err)
	}
	_, err = newManifestFile(badSuffix, fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), "should have .yaml suffix") {
		t.Fatalf("expect an error contains %q, but got %v", "does not exist", err)
	}
}
