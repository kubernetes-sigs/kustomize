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

	"k8s.io/kubectl/pkg/kustomize/constants"
	"k8s.io/kubectl/pkg/kustomize/types"
	"k8s.io/kubectl/pkg/kustomize/util/fs"
)

func TestWriteAndRead(t *testing.T) {
	kustomization := &types.Kustomization{
		NamePrefix: "prefix",
	}

	fsys := fs.MakeFakeFS()
	fsys.Create(constants.KustomizationFileName)
	mf, err := newKustomizationFile(constants.KustomizationFileName, fsys)
	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	if err := mf.write(kustomization); err != nil {
		t.Fatalf("Couldn't write kustomization file: %v\n", err)
	}

	content, err := mf.read()
	if err != nil {
		t.Fatalf("Couldn't read kustomization file: %v\n", err)
	}
	if !reflect.DeepEqual(kustomization, content) {
		t.Fatal("Read kustomization is different from written kustomization")
	}
}

func TestEmptyFile(t *testing.T) {
	fsys := fs.MakeFakeFS()
	_, err := newKustomizationFile("", fsys)
	if err == nil {
		t.Fatalf("Create kustomizationFile from empty filename should fail")
	}
}

func TestNewNotExist(t *testing.T) {
	badSuffix := "foo.bar"
	fakeFS := fs.MakeFakeFS()
	fakeFS.Mkdir(".", 0644)
	fakeFS.Create(badSuffix)
	_, err := newKustomizationFile(constants.KustomizationFileName, fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	contained := "Missing kustomization file"
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
	_, err = newKustomizationFile(constants.KustomizationFileName, fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
	_, err = newKustomizationFile(badSuffix, fakeFS)
	if err == nil {
		t.Fatalf("expect an error")
	}
	contained = "should have .yaml suffix"
	if !strings.Contains(err.Error(), contained) {
		t.Fatalf("expect an error contains %q, but got %v", contained, err)
	}
}
